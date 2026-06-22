package importer

import (
	"encoding/json"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
	session "github.com/autohttp/autohttp/session"
)

func parseFixture(data []byte) (*session.Session, error) {
	var raw struct {
		Exchanges []json.RawMessage `json:"exchanges"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	pbSession := &pb.RecordedSession{}
	for _, excRaw := range raw.Exchanges {
		var excRawMap map[string]json.RawMessage
		if err := json.Unmarshal(excRaw, &excRawMap); err != nil {
			return nil, err
		}

		exchange := &pb.HttpExchange{}

		if v, ok := excRawMap["id"]; ok {
			json.Unmarshal(v, &exchange.Id)
		}
		if v, ok := excRawMap["started_at"]; ok {
			json.Unmarshal(v, &exchange.StartedAt)
		}
		if v, ok := excRawMap["completed_at"]; ok {
			json.Unmarshal(v, &exchange.CompletedAt)
		}
		if v, ok := excRawMap["initiator"]; ok {
			json.Unmarshal(v, &exchange.Initiator)
		}
		if v, ok := excRawMap["request_body_complete"]; ok {
			json.Unmarshal(v, &exchange.RequestBodyComplete)
		}
		if v, ok := excRawMap["response_body_complete"]; ok {
			json.Unmarshal(v, &exchange.ResponseBodyComplete)
		}
		if v, ok := excRawMap["redirect_chain"]; ok {
			json.Unmarshal(v, &exchange.RedirectChain)
		}
		if v, ok := excRawMap["request"]; ok {
			exchange.Request = parseRequest(v)
		}
		if v, ok := excRawMap["response"]; ok {
			exchange.Response = parseResponse(v)
		}

		pbSession.Exchanges = append(pbSession.Exchanges, exchange)
	}

	return session.FromProto(pbSession), nil
}

func parseRequest(raw []byte) *pb.Request {
	var reqMap map[string]interface{}
	if err := json.Unmarshal(raw, &reqMap); err != nil {
		return nil
	}

	req := &pb.Request{}
	if v, ok := reqMap["method"].(string); ok {
		req.Method = v
	}
	if v, ok := reqMap["url"].(string); ok {
		req.Url = v
	}
	if v, ok := reqMap["body_type"].(string); ok {
		req.BodyType = v
	}
	if v, ok := reqMap["body"].(string); ok {
		req.Body = []byte(v)
	}

	if headers, ok := reqMap["headers"].([]interface{}); ok {
		for _, h := range headers {
			if m, ok := h.(map[string]interface{}); ok {
				key, _ := m["key"].(string)
				val, _ := m["value"].(string)
				req.Headers = append(req.Headers, &pb.Header{Key: key, Value: val})
			}
		}
	}

	if cookies, ok := reqMap["cookies"].([]interface{}); ok {
		for _, c := range cookies {
			if m, ok := c.(map[string]interface{}); ok {
				name, _ := m["name"].(string)
				val, _ := m["value"].(string)
				domain, _ := m["domain"].(string)
				path, _ := m["path"].(string)
				var expires int64
				if e, ok := m["expires"].(float64); ok {
					expires = int64(e)
				}
				httpOnly, _ := m["http_only"].(bool)
				secure, _ := m["secure"].(bool)
				sameSite, _ := m["same_site"].(string)
				req.Cookies = append(req.Cookies, &pb.CookieMutation{
					Name:     name,
					Value:    val,
					Domain:   domain,
					Path:     path,
					Expires:  expires,
					HttpOnly: httpOnly,
					Secure:   secure,
					SameSite: sameSite,
				})
			}
		}
	}

	return req
}

func parseResponse(raw []byte) *pb.Response {
	var respMap map[string]interface{}
	if err := json.Unmarshal(raw, &respMap); err != nil {
		return nil
	}

	resp := &pb.Response{}
	if v, ok := respMap["status"].(float64); ok {
		resp.Status = int32(v)
	}
	if v, ok := respMap["status_text"].(string); ok {
		resp.StatusText = v
	}
	if v, ok := respMap["body_type"].(string); ok {
		resp.BodyType = v
	}
	if v, ok := respMap["body"].(string); ok {
		resp.Body = []byte(v)
	}

	if headers, ok := respMap["headers"].([]interface{}); ok {
		for _, h := range headers {
			if m, ok := h.(map[string]interface{}); ok {
				key, _ := m["key"].(string)
				val, _ := m["value"].(string)
				resp.Headers = append(resp.Headers, &pb.Header{Key: key, Value: val})
			}
		}
	}

	if cookies, ok := respMap["set_cookies"].([]interface{}); ok {
		for _, c := range cookies {
			if m, ok := c.(map[string]interface{}); ok {
				name, _ := m["name"].(string)
				val, _ := m["value"].(string)
				domain, _ := m["domain"].(string)
				path, _ := m["path"].(string)
				var expires int64
				if e, ok := m["expires"].(float64); ok {
					expires = int64(e)
				}
				httpOnly, _ := m["http_only"].(bool)
				secure, _ := m["secure"].(bool)
				sameSite, _ := m["same_site"].(string)
				resp.SetCookies = append(resp.SetCookies, &pb.CookieMutation{
					Name:     name,
					Value:    val,
					Domain:   domain,
					Path:     path,
					Expires:  expires,
					HttpOnly: httpOnly,
					Secure:   secure,
					SameSite: sameSite,
				})
			}
		}
	}

	return resp
}
