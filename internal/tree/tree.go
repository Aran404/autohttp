package tree

import (
	"encoding/json"
	"fmt"
	"net/url"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(exchange *pb.HttpExchange) []*Tree {
	if exchange == nil {
		return nil
	}

	t := &Tree{
		Root:       &Node{Path: "", Type: "object"},
		ExchangeID: exchange.Id,
	}

	if exchange.Request != nil {
		parseRequest(t.Root, exchange)
	}
	if exchange.Response != nil {
		parseResponse(t.Root, exchange)
	}

	return []*Tree{t}
}

func (p *Parser) ParseSession(session *pb.RecordedSession) ([]*pb.ParsedTree, error) {
	return nil, nil
}

func parseRequest(parent *Node, exchange *pb.HttpExchange) {
	req := exchange.Request
	reqNode := &Node{Path: "request", Type: "object", Source: exchange.Id, SourceField: "request"}
	parent.Children = append(parent.Children, reqNode)

	reqNode.Children = append(reqNode.Children, &Node{
		Path: "method", Type: "string", RawValue: req.Method, Value: req.Method,
		Source: exchange.Id, SourceField: "request",
	})

	parseURL(reqNode, req.Url, exchange.Id)
	parseHeaders(reqNode, req.Headers, exchange.Id, "request")
	parseBody(reqNode, req.Body, req.BodyType, exchange.Id, "request")
}

func parseURL(parent *Node, rawURL string, sourceID string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		parent.Children = append(parent.Children, &Node{
			Path: "url", Type: "string", RawValue: rawURL, Value: rawURL,
			Source: sourceID, SourceField: "request.url",
		})
		return
	}

	urlNode := &Node{Path: "url", Type: "object", Source: sourceID, SourceField: "request.url"}
	parent.Children = append(parent.Children, urlNode)

	if u.Scheme != "" {
		urlNode.Children = append(urlNode.Children, &Node{
			Path: "scheme", Type: "string", RawValue: u.Scheme, Value: u.Scheme,
			Source: sourceID, SourceField: "request.url",
		})
	}
	if u.Host != "" {
		urlNode.Children = append(urlNode.Children, &Node{
			Path: "host", Type: "string", RawValue: u.Host, Value: u.Host,
			Source: sourceID, SourceField: "request.url",
		})
	}
	if u.Path != "" {
		urlNode.Children = append(urlNode.Children, &Node{
			Path: "path", Type: "string", RawValue: u.Path, Value: u.Path,
			Source: sourceID, SourceField: "request.url",
		})
	}
	if u.Fragment != "" {
		urlNode.Children = append(urlNode.Children, &Node{
			Path: "fragment", Type: "string", RawValue: u.Fragment, Value: u.Fragment,
			Source: sourceID, SourceField: "request.url",
		})
	}

	if q := u.Query(); len(q) > 0 {
		queryNode := &Node{Path: "query", Type: "object", Source: sourceID, SourceField: "request.url.query"}
		urlNode.Children = append(urlNode.Children, queryNode)
		for key, values := range q {
			for _, val := range values {
				queryNode.Children = append(queryNode.Children, &Node{
					Path: key, Type: "string", RawValue: val, Value: val,
					Source: sourceID, SourceField: "request.url.query",
				})
			}
		}
	}
}

func parseHeaders(parent *Node, headers []*pb.Header, sourceID string, side string) {
	if len(headers) == 0 {
		return
	}

	headersNode := &Node{
		Path: "headers", Type: "object", Source: sourceID, SourceField: side + ".headers",
	}
	parent.Children = append(parent.Children, headersNode)

	for _, h := range headers {
		headersNode.Children = append(headersNode.Children, &Node{
			Path: h.Key, Type: "string", RawValue: h.Value, Value: h.Value,
			Source: sourceID, SourceField: side + ".headers",
		})
	}
}

func parseBody(parent *Node, body []byte, bodyType string, sourceID string, side string) {
	bodyStr := string(body)
	if len(body) == 0 {
		return
	}

	bodyNode := &Node{
		Path: "body", Type: "object", Source: sourceID, SourceField: side + ".body",
	}
	parent.Children = append(parent.Children, bodyNode)

	if bodyType == "json" {
		var data any
		if err := json.Unmarshal(body, &data); err != nil {
			bodyNode.Type = "string"
			bodyNode.RawValue = bodyStr
			bodyNode.Value = bodyStr
			return
		}
		parseJSONValue(bodyNode, data, sourceID, side+".body")
	} else {
		bodyNode.Type = "string"
		bodyNode.RawValue = bodyStr
		bodyNode.Value = bodyStr
	}
}

func parseJSONValue(parent *Node, data any, sourceID string, sourceField string) {
	switch v := data.(type) {
	case map[string]any:
		parent.Type = "object"
		for key, val := range v {
			child := &Node{Path: key, Source: sourceID, SourceField: sourceField + "." + key}
			parseJSONValue(child, val, sourceID, sourceField+"."+key)
			parent.Children = append(parent.Children, child)
		}
	case []any:
		parent.Type = "array"
		for i, val := range v {
			idxStr := fmt.Sprintf("%d", i)
			child := &Node{Path: idxStr, Source: sourceID, SourceField: sourceField + "." + idxStr}
			parseJSONValue(child, val, sourceID, sourceField+"."+idxStr)
			parent.Children = append(parent.Children, child)
		}
	case string:
		parent.Type = "string"
		parent.RawValue = v
		parent.Value = v
	case float64:
		parent.Type = "number"
		parent.RawValue = fmt.Sprintf("%v", v)
		parent.Value = v
	case bool:
		parent.Type = "boolean"
		parent.RawValue = fmt.Sprintf("%v", v)
		parent.Value = v
	case nil:
		parent.Type = "null"
		parent.RawValue = "null"
		parent.Value = nil
	}
}

func parseResponse(parent *Node, exchange *pb.HttpExchange) {
	resp := exchange.Response
	respNode := &Node{Path: "response", Type: "object", Source: exchange.Id, SourceField: "response"}
	parent.Children = append(parent.Children, respNode)

	statusStr := fmt.Sprintf("%d", resp.Status)
	respNode.Children = append(respNode.Children, &Node{
		Path: "status", Type: "number", RawValue: statusStr, Value: int(resp.Status),
		Source: exchange.Id, SourceField: "response",
	})

	if resp.StatusText != "" {
		respNode.Children = append(respNode.Children, &Node{
			Path: "status_text", Type: "string", RawValue: resp.StatusText, Value: resp.StatusText,
			Source: exchange.Id, SourceField: "response",
		})
	}

	parseHeaders(respNode, resp.Headers, exchange.Id, "response")
	parseBody(respNode, resp.Body, resp.BodyType, exchange.Id, "response")
}
