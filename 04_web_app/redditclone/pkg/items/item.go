package items

type Author struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

type Comment struct {
	Created   string `json:"created"`
	Author    Author `json:"author"`
	Comment   string `json:"body"`
	CommentID string `json:"id"`
}

type Vote struct {
	UserID string `json:"user"`
	Vote   int8   `json:"vote"`
}

type Item struct {
	Score            int64           `json:"score"`
	Views            uint64          `json:"views"`
	Type             string          `json:"type"`
	Title            string          `json:"title"`
	Author           Author          `json:"author"`
	Category         string          `json:"category"`
	Text             string          `json:"text,omitempty"`
	URL              string          `json:"url,omitempty"`
	Votes            []Vote 		 `json:"votes"`
	Comments         []Comment      `json:"comments"`
	Created          string          `json:"created"`
	UpvotePercentage uint8           `json:"upvotePercentage"`
	PostID           string          `json:"id"`
}

type ItemsRepo interface {
	GetAll() ([]Item, error)
	GetByID(id string) (Item, error)
	Add(item *Item) (uint64, error)
	Update(newItem *Item) (bool, error)
	Delete(id string) (bool, error)
}
