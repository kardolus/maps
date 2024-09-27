package client_test

import (
	"fmt"
	"github.com/golang/mock/gomock"
	_ "github.com/golang/mock/mockgen/model"
	"github.com/kardolus/maps/client"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"testing"
)

//go:generate mockgen -destination=callermocks_test.go -package=client_test github.com/kardolus/maps/http Caller

const (
	apiKey             = "api-key"
	entity             = "New York"
	transformedEntity  = "New+York"
	singlePageResponse = `{
		"results": [
			{
				"name": "name",
				"geometry": {
					"location": {
						"lat": 1.0,
						"lng": 2.0
					}
				}
			}
		],
		"status": "OK"
	}`

	multiPageResponse = `{
		"results": [
			{
				"name": "name",
				"geometry": {
					"location": {
						"lat": 3.0,
						"lng": 4.0
					}
				}
			}
		],
		"next_page_token": "next-page-token",
		"status": "OK"
	}`
)

var (
	mockCtrl   *gomock.Controller
	mockCaller *MockCaller
	subject    *client.Client
)

func TestUnitClient(t *testing.T) {
	spec.Run(t, "Client Package Unit Tests", testClient, spec.Report(report.Terminal{}))
}

func testClient(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
		mockCtrl = gomock.NewController(t)
		mockCaller = NewMockCaller(mockCtrl)

		subject = client.New(mockCaller, apiKey)
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	when("FetchLocations()", func() {

		it("returns an error when the entity is empty", func() {
			_, err := subject.FetchLocations("", []string{}, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(client.ErrMissingEntity))
		})

		it("constructs the correct URL for a single entity", func() {
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(`{}`), nil)

			_, err := subject.FetchLocations(entity, []string{}, []string{})
			Expect(err).NotTo(HaveOccurred())
		})

		it("fetches locations with a single-page result", func() {
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(singlePageResponse), nil)

			result, err := subject.FetchLocations(entity, []string{"name"}, []string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveLen(1))
			Expect(result[0].Geometry.Location.Lat).To(Equal(1.0))
		})

		it("fetches locations with a multi-page result", func() {
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			// First page
			mockCaller.EXPECT().Get(expectedURL).Return([]byte(multiPageResponse), nil).Times(1)

			// Second page
			expectedNextPageURL := fmt.Sprintf(client.NextPageEndpoint, "next-page-token", apiKey)
			mockCaller.EXPECT().Get(expectedNextPageURL).Return([]byte(singlePageResponse), nil).Times(1)

			result, err := subject.FetchLocations(entity, []string{"name"}, []string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveLen(2))
			Expect(result[0].Geometry.Location.Lat).To(Equal(3.0))
			Expect(result[1].Geometry.Location.Lat).To(Equal(1.0))
		})

		it("filters locations based on contains and matches lists", func() {
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			responseWithDifferentNames := `{
				"results": [
					{"name": "name", "geometry": {"location": {"lat": 1.0, "lng": 2.0}}},
					{"name": "store", "geometry": {"location": {"lat": 3.0, "lng": 4.0}}},
					{"name": "Whole Foods Market", "geometry": {"location": {"lat": 5.0, "lng": 6.0}}}
				],
				"status": "OK"
			}`

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(responseWithDifferentNames), nil).Times(2)

			// Test contains filtering
			result, err := subject.FetchLocations(entity, []string{"Whole Foods"}, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Geometry.Location.Lat).To(Equal(5.0))

			// Test matches filtering
			result, err = subject.FetchLocations(entity, []string{}, []string{"store"})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Geometry.Location.Lat).To(Equal(3.0))
		})

		it("returns all locations when contains and matches lists are empty", func() {
			entity := "New York"
			transformedEntity := "New+York"
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			responseWithMultipleLocations := `{
				"results": [
					{"name": "Location A", "geometry": {"location": {"lat": 1.0, "lng": 2.0}}},
					{"name": "Location B", "geometry": {"location": {"lat": 3.0, "lng": 4.0}}}
				],
				"status": "OK"
			}`

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(responseWithMultipleLocations), nil)

			result, err := subject.FetchLocations(entity, []string{}, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(2)) // Expect all locations to be returned
			Expect(result[0].Geometry.Location.Lat).To(Equal(1.0))
			Expect(result[1].Geometry.Location.Lat).To(Equal(3.0))
		})

		it("filters locations based on case-insensitive contains and matches", func() {
			entity := "New York"
			transformedEntity := "New+York"
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			responseWithDifferentNames := `{
				"results": [
					{"name": "Whole Foods Market", "geometry": {"location": {"lat": 1.0, "lng": 2.0}}},
					{"name": "WHOLE FOODS market", "geometry": {"location": {"lat": 3.0, "lng": 4.0}}},
					{"name": "whole foods market", "geometry": {"location": {"lat": 5.0, "lng": 6.0}}}
				],
				"status": "OK"
			}`

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(responseWithDifferentNames), nil).Times(2)

			// Test contains filtering (case-insensitive)
			result, err := subject.FetchLocations(entity, []string{"whole foods market"}, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(3)) // Expect all three locations to match
			Expect(result[0].Geometry.Location.Lat).To(Equal(1.0))
			Expect(result[1].Geometry.Location.Lat).To(Equal(3.0))
			Expect(result[2].Geometry.Location.Lat).To(Equal(5.0))

			// Test matches filtering (case-insensitive)
			result, err = subject.FetchLocations(entity, []string{}, []string{"WHOLE FOODS MARKET"})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(3)) // Expect all three locations to match
			Expect(result[0].Geometry.Location.Lat).To(Equal(1.0))
			Expect(result[1].Geometry.Location.Lat).To(Equal(3.0))
			Expect(result[2].Geometry.Location.Lat).To(Equal(5.0))
		})

		it("handles no results returned by the API", func() {
			entity := "New York"
			transformedEntity := "New+York"
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			emptyResponse := `{
				"results": [],
				"status": "OK"
			}`

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(emptyResponse), nil)

			result, err := subject.FetchLocations(entity, []string{}, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		it("handles invalid or incomplete JSON response", func() {
			entity := "New York"
			transformedEntity := "New+York"
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			invalidJSONResponse := `{
				"results": [`

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(invalidJSONResponse), nil)

			result, err := subject.FetchLocations(entity, []string{}, []string{})
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})
}
