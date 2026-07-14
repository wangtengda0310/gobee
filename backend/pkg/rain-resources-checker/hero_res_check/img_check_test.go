package hero_res_check

import (
	"testing"
)

func TestGitlabAndLocalImgInExcel(t *testing.T) {
	rep, err := CheckGitlabAndLocalImgInExcel("../config/excel", "../../../client/Master/Card/Assets/Bundles/UI/Images/", "../../../client/Master/Card/Assets/Bundles/Prefabs/Prefab/")
	if err != nil {
		t.Errorf("CheckGitlabAndLocalImgInExcel() error = %v", err)
	}
	_ = rep
}
