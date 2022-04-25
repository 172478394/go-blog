package article

import (
    "fmt"
    "github.com/beevik/etree"
    "testing"
)

func TestGetMidi(t *testing.T) {
    doc := etree.NewDocument()
    if err := doc.ReadFromFile("../../static/pianomusic/001/0000002/0007445.musicxml"); err != nil {
        panic(err)
    }
    root := doc.SelectElement("score-partwise")
    workTitle := root.FindElement("./work/work-title")
    if workTitle == nil {
        work := etree.NewElement("work")
        workTitle = work.CreateElement("work-title")
        workTitle.CreateText("光年之外")
        root.InsertChildAt(0, work)
    }
    identification := root.FindElement("./identification")
    if identification == nil {
        identification = etree.NewElement("identification")
    }
    creator := root.FindElement("./identification/creator")
    if creator == nil {
        creator = etree.NewElement("creator")
        creator.CreateText("邓紫棋 GEM")
        creator.CreateAttr("type", "composer")
        identification.InsertChildAt(0, creator)
    }

    _ = doc.WriteToFile("../../static/pianomusic/001/0000002/0007445.musicxml")
    fmt.Println(creator.Text())
}
