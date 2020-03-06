/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package openapi

import (
	"github.com/qoobing/userd/src/api/openapi/createnewuser"
	"github.com/qoobing/userd/src/api/openapi/getcaptchaimage"
	"github.com/qoobing/userd/src/api/openapi/getuserinfo"
	"github.com/qoobing/userd/src/api/openapi/getverifycode"
	"github.com/qoobing/userd/src/api/openapi/login"
	"github.com/qoobing/userd/src/api/openapi/loginbyoauth20"
	"github.com/qoobing/userd/src/api/openapi/logout"
	"github.com/qoobing/userd/src/api/openapi/setUserInfoSelfDefined"
	"github.com/qoobing/userd/src/api/openapi/updateuserinfo"
	"github.com/qoobing/userd/src/api/openapi/updateuserpassword"
)

var (
	Login                  = login.Main
	Logout                 = logout.Main
	GetUserInfo            = getuserinfo.MainV1
	GetUserInfoV2          = getuserinfo.MainV2
	CreateNewUser          = createnewuser.Main
	GetVerifyCode          = getverifycode.MainV1
	GetVerifyCodeV2        = getverifycode.MainV2
	GetCaptchaImage        = getcaptchaimage.Main
	Loginbyoauth20         = loginbyoauth20.Main
	Updateuserpassword     = updateuserpassword.MainV1
	UpdateuserpasswordV2   = updateuserpassword.MainV2
	UpdateUserInfo         = updateuserinfo.Main
	SetUserInfoSelfDefined = setUserInfoSelfDefined.MainV2
)
