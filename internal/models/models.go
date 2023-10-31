package models

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

type ImageMeta struct {
	Name   string
	Type   string
	Height int
	Width  int
}

func CollectImageMeta(data []byte, name string) (*ImageMeta, error) {
	r := bytes.NewReader(data)
	//r.Seek(0, 0)

	imageData, imageType, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	b := imageData.Bounds()
	return &ImageMeta{
		Name:   name,
		Type:   imageType,
		Height: b.Max.Y,
		Width:  b.Max.X,
	}, nil
}

type InfoForThumbnail struct {
	Path string `json:"path"`
	Size int    `json:"size"`
}
