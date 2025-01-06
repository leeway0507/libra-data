package library_api

type BookItemsResponse struct {
	Docs []struct {
		Doc BookItemsDoc `json:"doc"`
	} `json:"docs"`
	LibNm     string           `json:"libNm"`
	NumFound  int              `json:"numFound"`
	PageNo    int              `json:"pageNo"`
	PageSize  int              `json:"pageSize"`
	Request   BookItemsRequest `json:"request"`
	ResultNum int              `json:"resultNum"`
}

type BookItemsDoc struct {
	AdditionSymbol string `json:"addition_symbol"`
	Authors        string `json:"authors"`
	BookImageURL   string `json:"bookImageURL"`
	Bookname       string `json:"bookname"`
	CallNumbers    []struct {
		CallNumber struct {
			BookCode          string `json:"book_code"`
			CopyCode          string `json:"copy_code"`
			SeparateShelfCode string `json:"separate_shelf_code"`
			SeparateShelfName string `json:"separate_shelf_name"`
			ShelfLocCode      string `json:"shelf_loc_code"`
			ShelfLocName      string `json:"shelf_loc_name"`
		} `json:"callNumber"`
	} `json:"callNumbers"`
	ClassNm         string `json:"class_nm"`
	ClassNo         string `json:"class_no"`
	ISBN13          string `json:"isbn13"`
	PublicationYear string `json:"publication_year"`
	Publisher       string `json:"publisher"`
	RegDate         string `json:"reg_date"`
	SetISBN13       string `json:"set_isbn13"`
	Vol             string `json:"vol"`
}

type BookItemsRequest struct {
	EndDt    string `json:"endDt"`
	LibCode  string `json:"libCode"`
	PageSize string `json:"pageSize"`
	StartDt  string `json:"startDt"`
}

type BookResp struct {
	Isbn        string
	ImageUrl    string
	Description string
	Source      string
	Url         string
}
