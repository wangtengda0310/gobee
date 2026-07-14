package hero_res_check

import (
	"testing"
)

func TestGitlabAndLocalVoicesInExcel(t *testing.T) {
	rep, err := CheckGitlabAndLocalVoicesInExcel("../config/excel", "../../../client/Master/Card/Audio/")
	if err != nil {
		t.Errorf("CheckGitlabAndLocalVoicesInExcel() error = %v", err)
	}
	_ = rep
}
