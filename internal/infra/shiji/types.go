package shiji

// ArticleListResponse 文章列表响应
type ArticleListResponse struct {
	Code  int       `json:"code"`
	Msg   string    `json:"msg"`
	Rows  []Article `json:"rows"`
	Total int64     `json:"total"`
}

// ArticleDetailResponse 文章详情响应
type ArticleDetailResponse struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data Article `json:"data"`
}

// Article 文章
type Article struct {
	ArticleId    int64  `json:"articleId,string"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Content      string `json:"content"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Status       int    `json:"status,string"`
	ViewCount    int    `json:"viewCount"`
	PublishTime  string `json:"publishTime"`
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpireIn    int64  `json:"expire_in"`
}
