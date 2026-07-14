package resource_utils

import (
	"encoding/json"
	"testing"
)

func TestLocal(t *testing.T) {
	info, err := GetFilesLocal("../../../client/Master/Card/Audio/Voice", "wav")

	if err != nil {
		t.Fatal(err)
	}
	indent, err := json.MarshalIndent(info, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(indent))
}

func TestGitlab(t *testing.T) {
	info, err := GetFilesGitlab("v0.0.8-pre-release", "Xcards/client", "Master/Card/Audio/Voice", "jMD7RXcbMLosvHoPZTbS", "wav")
	if err != nil {
		t.Fatal(err)
	}
	indent, err := json.MarshalIndent(info, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(indent))
}
