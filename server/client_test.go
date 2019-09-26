package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/mlinhard/exactly-index/search"
)

type testClient struct {
	t          *testing.T
	baseUrl    string
	maxContext int
	maxHits    int
}

type testClientResponse struct {
	client   *testClient
	response *SearchResponse
}

type testClientHit struct {
	response *testClientResponse
	hit      Hit
}

func (this *testClient) DocumentCount() int {
	return 1
}

func (this *testClient) Document(i int) *search.Document {
	request := new(DocumentRequest)
	response := new(DocumentResponse)
	request.DocumentIndex = &i
	err := this.postJson("/document", request, response)
	if err != nil {
		this.t.Error(err)
		return nil
	}
	doc := new(search.Document)
	doc.Content = response.Content
	doc.Id = response.DocumentId
	doc.Index = response.DocumentIndex
	return doc
}

func (this *testClient) search(pattern string) *testClientResponse {
	return this.searchBounded(pattern, this.maxContext, this.maxHits)
}

func (this *testClient) searchBounded(pattern string, maxContext, maxHits int) *testClientResponse {
	request := new(SearchRequest)
	response := new(SearchResponse)
	request.MaxContext = maxContext
	request.MaxHits = maxHits
	ofst := 0
	request.Offset = &ofst
	request.Pattern = []byte(pattern)
	err := this.postJson("/search", request, response)
	if err != nil {
		this.t.Error(err)
		return nil
	}
	return &testClientResponse{this, response}
}

func (this *testClient) postJson(relUrl string, request, response interface{}) error {
	reqData, err := json.Marshal(request)
	if err != nil {
		this.t.Errorf("marshalling: %v", err)
		return nil
	}
	var httpResp *http.Response
	httpResp, err = http.Post(this.baseUrl+relUrl, "application/json", bytes.NewReader(reqData))
	if err != nil {
		this.t.Errorf("http post: %v", err)
		return nil
	}
	respData, err := ioutil.ReadAll(httpResp.Body)
	fmt.Printf("received resp: %v", string(respData))
	err = json.Unmarshal(respData, response)
	if err != nil {
		this.t.Errorf("unmarshalling: %v", err)
	}
	return nil
}

func (this *testClientResponse) AssertHitCount(hitCount int) {
	if len(this.response.Hits) != hitCount {
		this.client.t.Errorf("Unexpected hit count %v (expected %v)", len(this.response.Hits), hitCount)
	}
}

func (this *testClientResponse) HitByDocumentId(docId string) *testClientHit {
	for _, hit := range this.response.Hits {
		if hit.DocumentId == docId {
			return &testClientHit{this, hit}
		}
	}
	return nil
}

func (this *testClientHit) AssertBefore(ctx string) *testClientHit {
	actualCtx := string(this.hit.ContextBefore)
	if actualCtx != ctx {
		this.response.client.t.Errorf("Expected before context %v and got %v", ctx, actualCtx)
	}
	return this
}

func (this *testClientHit) AssertAfter(ctx string) *testClientHit {
	actualCtx := string(this.hit.ContextAfter)
	if actualCtx != ctx {
		this.response.client.t.Errorf("Expected after context %v and got %v", ctx, actualCtx)
	}
	return this
}
