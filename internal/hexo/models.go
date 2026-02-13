package hexo

type Post struct {
	ID        string
	Title     string
	Date      string
	Updated   string
	Category  string
	Tags      []string
	BannerImg string
	IndexImg  string
	Content   string
	Math      bool
	Mermaid   bool
}
