package session

import (
	"time"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

type Session struct {
	ID        string
	TargetURL string
	StartedAt time.Time
	StoppedAt time.Time
	Exchanges []*Exchange
	Warnings  []string
}

type Exchange struct {
	ID          string
	Request     *Request
	Response    *Response
	StartedAt   time.Time
	CompletedAt time.Time
	Initiator   string
}

type Request struct {
	Method   string
	URL      string
	Headers  map[string]string
	Cookies  map[string]string
	Body     []byte
	BodyType string
}

type Response struct {
	Status     int
	StatusText string
	Headers    map[string]string
	SetCookies map[string]string
	Body       []byte
	BodyType   string
}

func FromProto(pb *pb.RecordedSession) *Session {
	if pb == nil {
		return nil
	}
	s := &Session{
		ID:        pb.Id,
		TargetURL: pb.TargetUrl,
		StartedAt: time.UnixMilli(pb.StartedAt),
		StoppedAt: time.UnixMilli(pb.StoppedAt),
		Warnings:  pb.Warnings,
	}
	for _, e := range pb.Exchanges {
		s.Exchanges = append(s.Exchanges, exchangeFromProto(e))
	}
	return s
}

func exchangeFromProto(pe *pb.HttpExchange) *Exchange {
	if pe == nil {
		return nil
	}
	e := &Exchange{
		ID:          pe.Id,
		StartedAt:   time.UnixMilli(pe.StartedAt),
		CompletedAt: time.UnixMilli(pe.CompletedAt),
		Initiator:   pe.Initiator,
	}
	if pe.Request != nil {
		e.Request = &Request{
			Method:   pe.Request.Method,
			URL:      pe.Request.Url,
			Headers:  make(map[string]string),
			Cookies:  make(map[string]string),
			Body:     pe.Request.Body,
			BodyType: pe.Request.BodyType,
		}
		for _, h := range pe.Request.Headers {
			e.Request.Headers[h.Key] = h.Value
		}
		for _, c := range pe.Request.Cookies {
			e.Request.Cookies[c.Name] = c.Value
		}
	}
	if pe.Response != nil {
		e.Response = &Response{
			Status:     int(pe.Response.Status),
			StatusText: pe.Response.StatusText,
			Headers:    make(map[string]string),
			SetCookies: make(map[string]string),
			Body:       pe.Response.Body,
			BodyType:   pe.Response.BodyType,
		}
		for _, h := range pe.Response.Headers {
			e.Response.Headers[h.Key] = h.Value
		}
		for _, c := range pe.Response.SetCookies {
			e.Response.SetCookies[c.Name] = c.Value
		}
	}
	return e
}

func (s *Session) ToProto() *pb.RecordedSession {
	panic("not implemented")
}
