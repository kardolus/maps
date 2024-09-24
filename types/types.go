package types

type Response struct {
	HtmlAttributions []interface{} `json:"html_attributions"`
	NextPageToken    string        `json:"next_page_token"`
	Results          []struct {
		BusinessStatus   string `json:"business_status"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			Viewport struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		Icon                string `json:"icon"`
		IconBackgroundColor string `json:"icon_background_color"`
		IconMaskBaseUri     string `json:"icon_mask_base_uri"`
		Name                string `json:"name"`
		OpeningHours        struct {
			OpenNow bool `json:"open_now"`
		} `json:"opening_hours"`
		Photos []struct {
			Height           int      `json:"height"`
			HtmlAttributions []string `json:"html_attributions"`
			PhotoReference   string   `json:"photo_reference"`
			Width            int      `json:"width"`
		} `json:"photos"`
		PlaceId  string `json:"place_id"`
		PlusCode struct {
			CompoundCode string `json:"compound_code"`
			GlobalCode   string `json:"global_code"`
		} `json:"plus_code"`
		PriceLevel       int      `json:"price_level,omitempty"`
		Rating           float64  `json:"rating"`
		Reference        string   `json:"reference"`
		Types            []string `json:"types"`
		UserRatingsTotal int      `json:"user_ratings_total"`
	} `json:"results"`
	Status string `json:"status"`
}
