package operation_with_headers

import (
	"context"
	"encoding/xml"
	"fmt"
	soap "github.com/justinclift-prvidr/soap-go"
)

// ClientOption configures a Client.
type ClientOption = soap.ClientOption

// Client is a SOAP client for this service.
type Client struct {
	*soap.Client
}

// NewClient creates a new SOAP client.
func NewClient(opts ...ClientOption) (*Client, error) {
	soapOpts := append([]soap.ClientOption{
		soap.WithEndpoint("http://example.com/headers"),
	}, opts...)
	soapClient, err := soap.NewClient(soapOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP client: %w", err)
	}
	return &Client{
		Client: soapClient,
	}, nil
}

// SubmitOrder executes the submitOrder SOAP operation.
func (c *Client) SubmitOrder(ctx context.Context, header *RequestHeaderWrapper, req *SubmitOrderWrapper, opts ...ClientOption) (*SubmitOrderResponseWrapper, error) {
	reqEnvelope, err := soap.NewEnvelope(soap.WithHeaderContent(header), soap.WithBody(req))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP envelope: %w", err)
	}
	respEnvelope, err := c.Call(ctx, "submitOrder", reqEnvelope, opts...)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %w", err)
	}
	var result SubmitOrderResponseWrapper
	if err := xml.Unmarshal(respEnvelope.Body.Content, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return &result, nil
}

// ReserveResourceResult bundles the typed SOAP Header and Body returned by ReserveResource.
type ReserveResourceResult struct {
	Header *ResponseHeaderWrapper
	Body   *ReserveResourceResponseWrapper
}

// ReserveResource executes the reserveResource SOAP operation.
func (c *Client) ReserveResource(ctx context.Context, header *RequestHeaderWrapper, req *ReserveResourceWrapper, opts ...ClientOption) (*ReserveResourceResult, error) {
	reqEnvelope, err := soap.NewEnvelope(soap.WithHeaderContent(header), soap.WithBody(req))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP envelope: %w", err)
	}
	respEnvelope, err := c.Call(ctx, "reserveResource", reqEnvelope, opts...)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %w", err)
	}
	var result ReserveResourceResponseWrapper
	if err := xml.Unmarshal(respEnvelope.Body.Content, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	if respEnvelope.Header == nil {
		return nil, fmt.Errorf("expected response header {%s}%s, envelope has no header", "http://example.com/headers", "ResponseHeader")
	}
	var respHeader ResponseHeaderWrapper
	foundHeader := false
	for _, entry := range respEnvelope.Header.Entries {
		if entry.XMLName.Space == "http://example.com/headers" && entry.XMLName.Local == "ResponseHeader" {
			if err := soap.UnmarshalHeaderEntry(entry, &respHeader); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response header: %w", err)
			}
			foundHeader = true
			break
		}
	}
	if !foundHeader {
		return nil, fmt.Errorf("expected response header {%s}%s, not found", "http://example.com/headers", "ResponseHeader")
	}
	return &ReserveResourceResult{Header: &respHeader, Body: &result}, nil
}
