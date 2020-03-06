/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package rbac

import (
	"github.com/qoobing/userd/src/api/rbac/basic/addRoleMap"
	"github.com/qoobing/userd/src/api/rbac/basic/checkPermission"
	"github.com/qoobing/userd/src/api/rbac/basic/createPrivilege"
	"github.com/qoobing/userd/src/api/rbac/basic/createRole"
	"github.com/qoobing/userd/src/api/rbac/basic/getPrivilegeDetail"
	"github.com/qoobing/userd/src/api/rbac/basic/getPrivilegeList"
	"github.com/qoobing/userd/src/api/rbac/basic/getRoleDetail"
	"github.com/qoobing/userd/src/api/rbac/basic/getRoleList"
	"github.com/qoobing/userd/src/api/rbac/template/createTemplateRole"
	"github.com/qoobing/userd/src/api/rbac/template/generateTemplateRoleTemplate"
	"github.com/qoobing/userd/src/api/rbac/template/getTemplateRoleDetail"
	"github.com/qoobing/userd/src/api/rbac/template/getTemplateRoleList"
	"github.com/qoobing/userd/src/api/rbac/template/updateTemplateRole"
)

var (
	CheckPermission    = checkPermission.Main
	CreateRole         = createRole.Main
	CreatePrivilege    = createPrivilege.Main
	AddRoleMap         = addRoleMap.Main
	GetPrivilegeList   = getPrivilegeList.Main
	GetPrivilegeDetail = getPrivilegeDetail.Main
	GetRoleList        = getRoleList.Main
	GetRoleDetail      = getRoleDetail.Main

	CreateTemplateRole           = createTemplateRole.Main
	GenerateTemplateRoleTemplate = generateTemplateRoleTemplate.Main
	GetTemplateRoleDetail        = getTemplateRoleDetail.Main
	GetTemplateRoleList          = getTemplateRoleList.Main
	UpdateTemplateRole           = updateTemplateRole.Main
)
