package xpath

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	path "github.com/antchfx/xpath"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type dataNode interface{}

type dataDocument interface {
	Parse(buf []byte) (dataNode, error)
	QueryAll(node dataNode, expr string) ([]dataNode, error)
	CreateXPathNavigator(node dataNode) path.NodeNavigator
	GetNodePath(node, relativeTo dataNode, sep string) string
	OutputXML(node dataNode) string
}

type Parser struct {
	Format              string            `toml:"-"`
	ProtobufMessageDef  string            `toml:"xpath_protobuf_file"`
	ProtobufMessageType string            `toml:"xpath_protobuf_type"`
	ProtobufImportPaths []string          `toml:"xpath_protobuf_import_paths"`
	PrintDocument       bool              `toml:"xpath_print_document"`
	AllowEmptySelection bool              `toml:"xpath_allow_empty_selection"`
	Configs             []Config          `toml:"xpath"`
	DefaultMetricName   string            `toml:"-"`
	DefaultTags         map[string]string `toml:"-"`
	Log                 telegraf.Logger   `toml:"-"`

	// Required for backward compatibility
	ConfigsXML     []Config `toml:"xml" deprecated:"1.23.1;use 'xpath' instead"`
	ConfigsJSON    []Config `toml:"xpath_json"`
	ConfigsMsgPack []Config `toml:"xpath_msgpack"`
	ConfigsProto   []Config `toml:"xpath_protobuf"`

	document dataDocument
}

// Config definition
// This should be replaced by the actual definition once
// the compatibitlity-code is removed.
// Please check plugins/parsers/registry.go for now.
type Config parsers.XPathConfig

func (p *Parser) Init() error {
	switch p.Format {
	case "", "xml":
		p.document = &xmlDocument{}

		// Required for backward compatibility
		if len(p.ConfigsXML) > 0 {
			p.Configs = append(p.Configs, p.ConfigsXML...)
			models.PrintOptionDeprecationNotice(telegraf.Warn, "parsers.xpath", "xml", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "2.0.0",
				Notice:    "use 'xpath' instead",
			})
		}
	case "xpath_json":
		p.document = &jsonDocument{}

		// Required for backward compatibility
		if len(p.ConfigsJSON) > 0 {
			p.Configs = append(p.Configs, p.ConfigsJSON...)
			models.PrintOptionDeprecationNotice(telegraf.Warn, "parsers.xpath", "xpath_json", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "2.0.0",
				Notice:    "use 'xpath' instead",
			})
		}
	case "xpath_msgpack":
		p.document = &msgpackDocument{}

		// Required for backward compatibility
		if len(p.ConfigsMsgPack) > 0 {
			p.Configs = append(p.Configs, p.ConfigsMsgPack...)
			models.PrintOptionDeprecationNotice(telegraf.Warn, "parsers.xpath", "xpath_msgpack", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "2.0.0",
				Notice:    "use 'xpath' instead",
			})
		}
	case "xpath_protobuf":
		pbdoc := protobufDocument{
			MessageDefinition: p.ProtobufMessageDef,
			MessageType:       p.ProtobufMessageType,
			ImportPaths:       p.ProtobufImportPaths,
			Log:               p.Log,
		}
		if err := pbdoc.Init(); err != nil {
			return err
		}
		p.document = &pbdoc

		// Required for backward compatibility
		if len(p.ConfigsProto) > 0 {
			p.Configs = append(p.Configs, p.ConfigsProto...)
			models.PrintOptionDeprecationNotice(telegraf.Warn, "parsers.xpath", "xpath_proto", telegraf.DeprecationInfo{
				Since:     "1.23.1",
				RemovalIn: "2.0.0",
				Notice:    "use 'xpath' instead",
			})
		}
	default:
		return fmt.Errorf("unknown data-format %q for xpath parser", p.Format)
	}

	// Make sure we do have a metric name
	if p.DefaultMetricName == "" {
		return errors.New("missing default metric name")
	}

	return nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	t := time.Now()

	// Parse the XML
	doc, err := p.document.Parse(buf)
	if err != nil {
		return nil, err
	}
	if p.PrintDocument {
		p.Log.Debugf("XML document equivalent: %q", p.document.OutputXML(doc))
	}

	// Queries
	metrics := make([]telegraf.Metric, 0)
	p.Log.Debugf("Number of configs: %d", len(p.Configs))
	for _, config := range p.Configs {
		if len(config.Selection) == 0 {
			config.Selection = "/"
		}
		selectedNodes, err := p.document.QueryAll(doc, config.Selection)
		if err != nil {
			return nil, err
		}
		if len(selectedNodes) < 1 || selectedNodes[0] == nil {
			p.debugEmptyQuery("metric selection", doc, config.Selection)
			if !p.AllowEmptySelection {
				return metrics, fmt.Errorf("cannot parse with empty selection node")
			}
		}
		p.Log.Debugf("Number of selected metric nodes: %d", len(selectedNodes))

		for _, selected := range selectedNodes {
			m, err := p.parseQuery(t, doc, selected, config)
			if err != nil {
				return metrics, err
			}

			metrics = append(metrics, m)
		}
	}

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	switch len(metrics) {
	case 0:
		return nil, nil
	case 1:
		return metrics[0], nil
	default:
		return metrics[0], fmt.Errorf("cannot parse line with multiple (%d) metrics", len(metrics))
	}
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) parseQuery(starttime time.Time, doc, selected dataNode, config Config) (telegraf.Metric, error) {
	var timestamp time.Time
	var metricname string

	// Determine the metric name. If a query was specified, use the result of this query and the default metric name
	// otherwise.
	metricname = p.DefaultMetricName
	if len(config.MetricQuery) > 0 {
		v, err := p.executeQuery(doc, selected, config.MetricQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to query metric name: %v", err)
		}
		var ok bool
		if metricname, ok = v.(string); !ok {
			if v == nil {
				p.Log.Infof("Hint: Empty metric-name-node. If you wanted to set a constant please use `metric_name = \"'name'\"`.")
			}
			return nil, fmt.Errorf("failed to query metric name: query result is of type %T not 'string'", v)
		}
	}

	// By default take the time the parser was invoked and override the value
	// with the queried timestamp if an expresion was specified.
	timestamp = starttime
	if len(config.Timestamp) > 0 {
		v, err := p.executeQuery(doc, selected, config.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to query timestamp: %v", err)
		}
		switch v := v.(type) {
		case string:
			// Parse the string with the given format or assume the string to contain
			// a unix timestamp in seconds if no format is given.
			if len(config.TimestampFmt) < 1 || strings.HasPrefix(config.TimestampFmt, "unix") {
				var nanoseconds int64

				t, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse unix timestamp: %v", err)
				}

				switch config.TimestampFmt {
				case "unix_ns":
					nanoseconds = int64(t)
				case "unix_us":
					nanoseconds = int64(t * 1e3)
				case "unix_ms":
					nanoseconds = int64(t * 1e6)
				default:
					nanoseconds = int64(t * 1e9)
				}
				timestamp = time.Unix(0, nanoseconds)
			} else {
				timestamp, err = time.Parse(config.TimestampFmt, v)
				if err != nil {
					return nil, fmt.Errorf("failed to query timestamp format: %v", err)
				}
			}
		case float64:
			// Assume the value to contain a timestamp in seconds and fractions thereof.
			timestamp = time.Unix(0, int64(v*1e9))
		case nil:
			// No timestamp found. Just ignore the time and use "starttime"
		default:
			return nil, fmt.Errorf("unknown format '%T' for timestamp query '%v'", v, config.Timestamp)
		}
	}

	// Query tags and add default ones
	tags := make(map[string]string)
	for name, query := range config.Tags {
		// Execute the query and cast the returned values into strings
		v, err := p.executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query tag '%s': %v", name, err)
		}
		switch v := v.(type) {
		case string:
			tags[name] = v
		case bool:
			tags[name] = strconv.FormatBool(v)
		case float64:
			tags[name] = strconv.FormatFloat(v, 'G', -1, 64)
		case nil:
			continue
		default:
			return nil, fmt.Errorf("unknown format '%T' for tag '%s'", v, name)
		}
	}

	// Handle the tag batch definitions if any.
	if len(config.TagSelection) > 0 {
		tagnamequery := "name()"
		tagvaluequery := "."
		if len(config.TagNameQuery) > 0 {
			tagnamequery = config.TagNameQuery
		}
		if len(config.TagValueQuery) > 0 {
			tagvaluequery = config.TagValueQuery
		}

		// Query all tags
		selectedTagNodes, err := p.document.QueryAll(selected, config.TagSelection)
		if err != nil {
			return nil, err
		}
		p.Log.Debugf("Number of selected tag nodes: %d", len(selectedTagNodes))
		if len(selectedTagNodes) > 0 && selectedTagNodes[0] != nil {
			for _, selectedtag := range selectedTagNodes {
				n, err := p.executeQuery(doc, selectedtag, tagnamequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query tag name with query '%s': %v", tagnamequery, err)
				}
				name, ok := n.(string)
				if !ok {
					return nil, fmt.Errorf("failed to query tag name with query '%s': result is not a string (%v)", tagnamequery, n)
				}
				v, err := p.executeQuery(doc, selectedtag, tagvaluequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query tag value for '%s': %v", name, err)
				}

				if config.TagNameExpand {
					p := p.document.GetNodePath(selectedtag, selected, "_")
					if len(p) > 0 {
						name = p + "_" + name
					}
				}

				// Check if field name already exists and if so, append an index number.
				if _, ok := tags[name]; ok {
					for i := 1; ; i++ {
						p := name + "_" + strconv.Itoa(i)
						if _, ok := tags[p]; !ok {
							name = p
							break
						}
					}
				}

				// Convert the tag to be a string
				s, err := internal.ToString(v)
				if err != nil {
					return nil, fmt.Errorf("failed to query tag value for '%s': result is not a string (%v)", name, v)
				}
				tags[name] = s
			}
		} else {
			p.debugEmptyQuery("tag selection", selected, config.TagSelection)
		}
	}

	for name, v := range p.DefaultTags {
		tags[name] = v
	}

	// Query fields
	fields := make(map[string]interface{})
	for name, query := range config.FieldsInt {
		// Execute the query and cast the returned values into integers
		v, err := p.executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query field (int) '%s': %v", name, err)
		}
		switch v := v.(type) {
		case string:
			fields[name], err = strconv.ParseInt(v, 10, 54)
			if err != nil {
				return nil, fmt.Errorf("failed to parse field (int) '%s': %v", name, err)
			}
		case bool:
			fields[name] = int64(0)
			if v {
				fields[name] = int64(1)
			}
		case float64:
			fields[name] = int64(v)
		case nil:
			continue
		default:
			return nil, fmt.Errorf("unknown format '%T' for field (int) '%s'", v, name)
		}
	}

	for name, query := range config.Fields {
		// Execute the query and store the result in fields
		v, err := p.executeQuery(doc, selected, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query field '%s': %v", name, err)
		}
		fields[name] = v
	}

	// Handle the field batch definitions if any.
	if len(config.FieldSelection) > 0 {
		fieldnamequery := "name()"
		fieldvaluequery := "."
		if len(config.FieldNameQuery) > 0 {
			fieldnamequery = config.FieldNameQuery
		}
		if len(config.FieldValueQuery) > 0 {
			fieldvaluequery = config.FieldValueQuery
		}

		// Query all fields
		selectedFieldNodes, err := p.document.QueryAll(selected, config.FieldSelection)
		if err != nil {
			return nil, err
		}
		p.Log.Debugf("Number of selected field nodes: %d", len(selectedFieldNodes))
		if len(selectedFieldNodes) > 0 && selectedFieldNodes[0] != nil {
			for _, selectedfield := range selectedFieldNodes {
				n, err := p.executeQuery(doc, selectedfield, fieldnamequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query field name with query '%s': %v", fieldnamequery, err)
				}
				name, ok := n.(string)
				if !ok {
					return nil, fmt.Errorf("failed to query field name with query '%s': result is not a string (%v)", fieldnamequery, n)
				}
				v, err := p.executeQuery(doc, selectedfield, fieldvaluequery)
				if err != nil {
					return nil, fmt.Errorf("failed to query field value for '%s': %v", name, err)
				}

				if config.FieldNameExpand {
					p := p.document.GetNodePath(selectedfield, selected, "_")
					if len(p) > 0 {
						name = p + "_" + name
					}
				}

				// Check if field name already exists and if so, append an index number.
				if _, ok := fields[name]; ok {
					for i := 1; ; i++ {
						p := name + "_" + strconv.Itoa(i)
						if _, ok := fields[p]; !ok {
							name = p
							break
						}
					}
				}

				fields[name] = v
			}
		} else {
			p.debugEmptyQuery("field selection", selected, config.FieldSelection)
		}
	}

	return metric.New(metricname, tags, fields, timestamp), nil
}

func (p *Parser) executeQuery(doc, selected dataNode, query string) (r interface{}, err error) {
	// Check if the query is relative or absolute and set the root for the query
	root := selected
	if strings.HasPrefix(query, "/") {
		root = doc
	}

	// Compile the query
	expr, err := path.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("failed to compile query '%s': %v", query, err)
	}

	// Evaluate the compiled expression and handle returned node-iterators
	// separately. Those iterators will be returned for queries directly
	// referencing a node (value or attribute).
	n := expr.Evaluate(p.document.CreateXPathNavigator(root))
	if iter, ok := n.(*path.NodeIterator); ok {
		// We got an iterator, so take the first match and get the referenced
		// property. This will always be a string.
		if iter.MoveNext() {
			r = iter.Current().Value()
		}
	} else {
		r = n
	}

	return r, nil
}

func splitLastPathElement(query string) []string {
	// This is a rudimentary xpath-parser that splits the path
	// into the last path element and the remaining path-part.
	// The last path element is then further splitted into
	// parts such as attributes or selectors. Each returned
	// element is a full path!

	// Nothing left
	if query == "" || query == "/" || query == "//" || query == "." {
		return []string{}
	}

	seperatorIdx := strings.LastIndex(query, "/")
	if seperatorIdx < 0 {
		query = "./" + query
		seperatorIdx = 1
	}

	// For double slash we want to split at the first slash
	if seperatorIdx > 0 && query[seperatorIdx-1] == byte('/') {
		seperatorIdx--
	}

	base := query[:seperatorIdx]
	if base == "" {
		base = "/"
	}

	elements := make([]string, 1)
	elements[0] = base

	offset := seperatorIdx
	if i := strings.Index(query[offset:], "::"); i >= 0 {
		// Check for axis operator
		offset += i
		elements = append(elements, query[:offset]+"::*")
	}

	if i := strings.Index(query[offset:], "["); i >= 0 {
		// Check for predicates
		offset += i
		elements = append(elements, query[:offset])
	} else if i := strings.Index(query[offset:], "@"); i >= 0 {
		// Check for attributes
		offset += i
		elements = append(elements, query[:offset])
	}

	return elements
}

func (p *Parser) debugEmptyQuery(operation string, root dataNode, initialquery string) {
	if p.Log == nil {
		return
	}

	query := initialquery

	// We already know that the
	p.Log.Debugf("got 0 nodes for query %q in %s", query, operation)
	for {
		parts := splitLastPathElement(query)
		if len(parts) < 1 {
			return
		}
		for i := len(parts) - 1; i >= 0; i-- {
			q := parts[i]
			nodes, err := p.document.QueryAll(root, q)
			if err != nil {
				p.Log.Debugf("executing query %q in %s failed: %v", q, operation, err)
				return
			}
			p.Log.Debugf("got %d nodes for query %q in %s", len(nodes), q, operation)
			if len(nodes) > 0 && nodes[0] != nil {
				return
			}
			query = parts[0]
		}
	}
}

func init() {
	// Register all variants
	parsers.Add("xml",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xml",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
	parsers.Add("xpath_json",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xpath_json",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
	parsers.Add("xpath_msgpack",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xpath_msgpack",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
	parsers.Add("xpath_protobuf",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				Format:            "xpath_protobuf",
				DefaultMetricName: defaultMetricName,
			}
		},
	)
}

// InitFromConfig is a compatibility function to construct the parser the old way
func (p *Parser) InitFromConfig(config *parsers.Config) error {
	p.Format = config.DataFormat
	if p.Format == "xpath_protobuf" {
		p.ProtobufMessageDef = config.XPathProtobufFile
		p.ProtobufMessageType = config.XPathProtobufType
	}
	p.PrintDocument = config.XPathPrintDocument
	p.DefaultMetricName = config.MetricName
	p.DefaultTags = config.DefaultTags

	// Convert the config formats which is a one-to-one copy
	if len(config.XPathConfig) > 0 {
		p.Configs = make([]Config, 0, len(config.XPathConfig))
		for _, cfg := range config.XPathConfig {
			config := Config(cfg)
			p.Configs = append(p.Configs, config)
		}
	}

	return p.Init()
}
