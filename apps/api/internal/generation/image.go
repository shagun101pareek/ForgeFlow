package generation

import "encoding/base64"

func applyProjectImage(spec Spec, dataURL string) Spec {
	if dataURL == "" || len(spec.Pages) == 0 {
		return spec
	}
	for i, section := range spec.Pages[0].Sections {
		if section.Type == "hero" {
			spec.Pages[0].Sections[i].Image = dataURL
			return spec
		}
	}
	return spec
}

func stripImages(spec Spec) Spec {
	for i := range spec.Pages {
		for j := range spec.Pages[i].Sections {
			spec.Pages[i].Sections[j].Image = ""
		}
	}
	return spec
}

func imageDataURL(contentType string, data []byte) string {
	if contentType == "" || len(data) == 0 {
		return ""
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func DetectImageType(data []byte) string {
	if len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return "image/jpeg"
	}
	if len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a") {
		return "image/gif"
	}
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	return ""
}
