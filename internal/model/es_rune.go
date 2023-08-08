package model

type ESRune struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
	Description string `json:"description,omitempty"`
	Plaintext   string `json:"plaintext,omitempty"`
	IconPath    string `json:"iconPath,omitempty"`
	Tooltip     string `json:"tooltip,omitempty"`
	SlotLabel   string `json:"slotLabel,omitempty"`
	StyleName   string `json:"styleName,omitempty"`
	Maps        string `json:"maps,omitempty"`
	Types       string `json:"types,omitempty"`
	Version     string `json:"version,omitempty"`
	FileTime    string `json:"fileTime,omitempty"`
	Platform    string `json:"platform"`
}

func (e *ESRune) GetMapping() string {
	return `
{
    "mappings": {
        "properties": {
            "name": {
                "type": "text",
                "analyzer": "ik_smart"
            },
            "keywords": {
                "type": "text",
                "analyzer": "ik_smart"
            },
            "plaintext": {
                "type": "text",
                "analyzer": "ik_smart"
            },
            "description": {
                "type": "text",
                "analyzer": "ik_smart"
            },
            "iconPath": {
                "type": "keyword"
            },
            "tooltip": {
                "type": "text",
                "analyzer": "ik_smart"
            },
            "slotLabel": {
                "type": "keyword"
            },
            "styleName": {
                "type": "keyword"
            },
            "maps": {
                "type": "keyword"
            },
            "types": {
                "type": "keyword"
            },
            "fileTime": {
                "type": "keyword"
            },
            "platform": {
                "type": "keyword"
            },
            "version": {
                "type": "keyword"
            }
        }
    }
}
`
}

func (e *ESRune) GetIndexName() string {
	return "lol_rune"
}
