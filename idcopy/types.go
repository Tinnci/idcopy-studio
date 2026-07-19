package idcopy

// SizeMM is a physical size in millimetres.
type SizeMM struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type PaperSpec struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Group  string  `json:"group"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type DocumentSpec struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Region   string  `json:"region"`
	Standard string  `json:"standard"`
	Kind     string  `json:"kind"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Notes    string  `json:"notes,omitempty"`
}

type Catalog struct {
	Papers    []PaperSpec    `json:"papers"`
	Documents []DocumentSpec `json:"documents"`
}

type ImageInput struct {
	DataURL  string `json:"dataURL"`
	Rotation int    `json:"rotation"`
}

type PDFOptions struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Subject  string `json:"subject"`
	Keywords string `json:"keywords"`
}

type LayoutRequest struct {
	PaperID        string     `json:"paperID"`
	CustomPaper    SizeMM     `json:"customPaper"`
	DocumentID     string     `json:"documentID"`
	CustomDocument SizeMM     `json:"customDocument"`
	Orientation    string     `json:"orientation"`
	Arrangement    string     `json:"arrangement"`
	Sides          string     `json:"sides"`
	Copies         int        `json:"copies"`
	ScalePercent   float64    `json:"scalePercent"`
	MarginMM       float64    `json:"marginMM"`
	GapMM          float64    `json:"gapMM"`
	Fit            string     `json:"fit"`
	CutMarks       bool       `json:"cutMarks"`
	Front          ImageInput `json:"front"`
	Back           ImageInput `json:"back"`
	PDF            PDFOptions `json:"pdf"`
}

type PlacedItem struct {
	Page   int     `json:"page"`
	Copy   int     `json:"copy"`
	Side   string  `json:"side"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type LayoutPlan struct {
	Paper       SizeMM       `json:"paper"`
	Document    SizeMM       `json:"document"`
	Orientation string       `json:"orientation"`
	Pages       int          `json:"pages"`
	Items       []PlacedItem `json:"items"`
}

type ExportResult struct {
	Filename   string `json:"filename"`
	DataBase64 string `json:"dataBase64"`
	Pages      int    `json:"pages"`
	Bytes      int    `json:"bytes"`
}

type SaveResult struct {
	Path     string `json:"path"`
	Pages    int    `json:"pages"`
	Bytes    int    `json:"bytes"`
	Canceled bool   `json:"canceled"`
}
