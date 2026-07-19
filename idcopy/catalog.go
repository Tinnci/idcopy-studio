package idcopy

import "sort"

var papers = []PaperSpec{
	{ID: "iso-a3", Name: "A3", Group: "ISO 216", Width: 297, Height: 420},
	{ID: "iso-a4", Name: "A4", Group: "ISO 216", Width: 210, Height: 297},
	{ID: "iso-a5", Name: "A5", Group: "ISO 216", Width: 148, Height: 210},
	{ID: "iso-a6", Name: "A6", Group: "ISO 216", Width: 105, Height: 148},
	{ID: "iso-b4", Name: "B4", Group: "ISO 216", Width: 250, Height: 353},
	{ID: "iso-b5", Name: "B5", Group: "ISO 216", Width: 176, Height: 250},
	{ID: "jis-b5", Name: "JIS B5", Group: "JIS P 0138", Width: 182, Height: 257},
	{ID: "ansi-letter", Name: "US Letter", Group: "ANSI", Width: 215.9, Height: 279.4},
	{ID: "ansi-legal", Name: "US Legal", Group: "ANSI", Width: 215.9, Height: 355.6},
	{ID: "ansi-tabloid", Name: "Tabloid", Group: "ANSI", Width: 279.4, Height: 431.8},
	{ID: "na-executive", Name: "Executive", Group: "North America", Width: 184.15, Height: 266.7},
	{ID: "cn-16k", Name: "16K", Group: "China", Width: 195, Height: 270},
	{ID: "jp-hagaki", Name: "Hagaki", Group: "Japan", Width: 100, Height: 148},
	{ID: "photo-4x6", Name: "4 × 6 in", Group: "Photo", Width: 101.6, Height: 152.4},
	{ID: "photo-5x7", Name: "5 × 7 in", Group: "Photo", Width: 127, Height: 177.8},
}

var documents = []DocumentSpec{
	{ID: "iso-id-1", Name: "ISO ID-1 card", Region: "International", Standard: "ISO/IEC 7810 ID-1", Kind: "card", Width: 85.60, Height: 53.98, Notes: "Most bank cards, driving licences and modern national ID cards."},
	{ID: "iso-id-2", Name: "ISO ID-2", Region: "International", Standard: "ISO/IEC 7810 ID-2", Kind: "document", Width: 105, Height: 74},
	{ID: "iso-id-3", Name: "ISO ID-3", Region: "International", Standard: "ISO/IEC 7810 ID-3", Kind: "passport", Width: 125, Height: 88},
	{ID: "iso-id-000", Name: "ISO ID-000", Region: "International", Standard: "ISO/IEC 7810 ID-000", Kind: "card", Width: 25, Height: 15},
	{ID: "icao-td1", Name: "ICAO TD1", Region: "International", Standard: "ICAO Doc 9303 TD1", Kind: "travel-document", Width: 85.60, Height: 54},
	{ID: "icao-td2", Name: "ICAO TD2", Region: "International", Standard: "ICAO Doc 9303 TD2", Kind: "travel-document", Width: 105, Height: 74},
	{ID: "icao-td3", Name: "ICAO TD3 passport page", Region: "International", Standard: "ICAO Doc 9303 TD3", Kind: "passport", Width: 125, Height: 88},
	{ID: "cn-resident-id", Name: "Mainland China resident ID", Region: "China", Standard: "GB 11643 / ID-1 form", Kind: "national-id", Width: 85.60, Height: 54},
	{ID: "hk-id", Name: "Hong Kong identity card", Region: "Hong Kong SAR", Standard: "ID-1 form", Kind: "national-id", Width: 85.60, Height: 54},
	{ID: "mo-id", Name: "Macao resident identity card", Region: "Macao SAR", Standard: "ID-1 form", Kind: "national-id", Width: 85.60, Height: 54},
	{ID: "tw-id", Name: "Taiwan national ID card", Region: "Taiwan", Standard: "ID-1 form", Kind: "national-id", Width: 85.60, Height: 54},
	{ID: "jp-my-number", Name: "Japan My Number card", Region: "Japan", Standard: "JIS X 6301 / ID-1", Kind: "national-id", Width: 85.60, Height: 53.98},
	{ID: "kr-resident-card", Name: "Korea resident registration card", Region: "South Korea", Standard: "ID-1 form", Kind: "national-id", Width: 85.60, Height: 54},
	{ID: "sg-nric", Name: "Singapore NRIC", Region: "Singapore", Standard: "ID-1 form", Kind: "national-id", Width: 85.60, Height: 53.98},
	{ID: "my-mykad", Name: "Malaysia MyKad", Region: "Malaysia", Standard: "ISO/IEC 7810 ID-1", Kind: "national-id", Width: 85.60, Height: 53.98},
	{ID: "th-national-id", Name: "Thailand national ID", Region: "Thailand", Standard: "ID-1 form", Kind: "national-id", Width: 85.60, Height: 54},
	{ID: "id-e-ktp", Name: "Indonesia e-KTP", Region: "Indonesia", Standard: "ISO/IEC 7810 ID-1", Kind: "national-id", Width: 85.60, Height: 53.98},
	{ID: "in-aadhaar-pvc", Name: "India Aadhaar PVC", Region: "India", Standard: "ISO/IEC 7810 ID-1", Kind: "national-id", Width: 85.60, Height: 53.98},
	{ID: "ph-philsys", Name: "Philippines PhilID", Region: "Philippines", Standard: "ISO/IEC 7810 ID-1", Kind: "national-id", Width: 85.60, Height: 53.98},
	{ID: "eu-id-card", Name: "EU identity card", Region: "European Union", Standard: "Regulation (EU) 2019/1157 / ID-1", Kind: "national-id", Width: 85.60, Height: 53.98},
	{ID: "us-passport-card", Name: "US passport card", Region: "United States", Standard: "ISO/IEC 7810 ID-1", Kind: "passport-card", Width: 85.60, Height: 53.98},
	{ID: "na-driving-licence", Name: "North American driving licence", Region: "United States / Canada", Standard: "AAMVA / ID-1 form", Kind: "driving-licence", Width: 85.60, Height: 53.98, Notes: "Issuer tolerances and older cards may vary."},
	{ID: "au-driving-licence", Name: "Australian driving licence", Region: "Australia", Standard: "Common ID-1 form", Kind: "driving-licence", Width: 85.60, Height: 53.98, Notes: "Formats vary by state and issue generation."},
	{ID: "nz-driving-licence", Name: "New Zealand driver licence", Region: "New Zealand", Standard: "ID-1 form", Kind: "driving-licence", Width: 86, Height: 54},
	{ID: "latam-id-1", Name: "Latin America modern ID card", Region: "Latin America", Standard: "Common ID-1 form", Kind: "national-id", Width: 85.60, Height: 53.98, Notes: "Use custom size for legacy or country-specific documents."},
}

func GetCatalog() Catalog {
	result := Catalog{
		Papers:    append([]PaperSpec(nil), papers...),
		Documents: append([]DocumentSpec(nil), documents...),
	}
	sort.SliceStable(result.Papers, func(i, j int) bool {
		if result.Papers[i].Group == result.Papers[j].Group {
			return result.Papers[i].Name < result.Papers[j].Name
		}
		return result.Papers[i].Group < result.Papers[j].Group
	})
	sort.SliceStable(result.Documents, func(i, j int) bool {
		if result.Documents[i].Region == result.Documents[j].Region {
			return result.Documents[i].Name < result.Documents[j].Name
		}
		return result.Documents[i].Region < result.Documents[j].Region
	})
	return result
}

func findPaper(id string) (PaperSpec, bool) {
	for _, paper := range papers {
		if paper.ID == id {
			return paper, true
		}
	}
	return PaperSpec{}, false
}

func findDocument(id string) (DocumentSpec, bool) {
	for _, document := range documents {
		if document.ID == id {
			return document, true
		}
	}
	return DocumentSpec{}, false
}
