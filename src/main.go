/***********************************************************************
// Copyright qoobing.com @2017 The source code.
// Copyright (c) 2009-2016 The Bitcoin Core developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package main

import (
	"github.com/labstack/echo"
	"github.com/qoobing/userd/src/api/corporation"
	"github.com/qoobing/userd/src/api/innerapi"
	"github.com/qoobing/userd/src/api/oauth20"
	"github.com/qoobing/userd/src/api/openapi"
	"github.com/qoobing/userd/src/api/rbac"
	"github.com/qoobing/userd/src/config"
	"github.com/qoobing/userd/src/model"
	"qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/gls"
	"qoobing.com/utillib.golang/xyz"
	"yqtc.com/log"
)

func main() {
	model.InitDatabase()
	config.InitEureka()
	e := echo.New()
	e.HTTPErrorHandler = func(e error, context echo.Context) {
		request := context.Request()
		log.Noticef("http error, %s, method:%s, uri:%s", e.Error(), request.Method, request.URL)
	}
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := apicontext.New(c, config.Config())
			// set logid
			req := cc.Request()
			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = xyz.GetRandomString(12)
			}
			gls.SetGlsValue("logid", id)

			// do next
			return h(cc)
		}
	})
	e.OPTIONS("*", func(cc echo.Context) error {
		return nil
	})

	e.GET("/health", config.Health)
	e.POST("/health", config.Health)
	e.GET("/openapi/login", openapi.Login)
	e.POST("/openapi/login", openapi.Login)
	e.GET("/openapi/logout", openapi.Logout)
	e.POST("/openapi/logout", openapi.Logout)
	e.GET("/openapi/getuserinfo", openapi.GetUserInfo)
	e.POST("/openapi/getuserinfo", openapi.GetUserInfo)
	e.GET("/openapi/updateUserInfo", openapi.UpdateUserInfo)
	e.POST("/openapi/updateUserInfo", openapi.UpdateUserInfo)
	e.GET("/openapi/getUserInfo", openapi.GetUserInfoV2)
	e.POST("/openapi/getUserInfo", openapi.GetUserInfoV2)
	e.GET("/openapi/createnewuser", openapi.CreateNewUser)
	e.POST("/openapi/createnewuser", openapi.CreateNewUser)
	e.GET("/openapi/getverifycode", openapi.GetVerifyCode)
	e.POST("/openapi/getverifycode", openapi.GetVerifyCode)
	e.GET("/openapi/getVerifyCode", openapi.GetVerifyCodeV2)
	e.POST("/openapi/getVerifyCode", openapi.GetVerifyCodeV2)
	e.GET("/openapi/getcaptchaimage", openapi.GetCaptchaImage)
	e.POST("/openapi/getcaptchaimage", openapi.GetCaptchaImage)
	e.GET("/openapi/loginbyoauth20", openapi.Loginbyoauth20)
	e.POST("/openapi/loginbyoauth20", openapi.Loginbyoauth20)
	e.GET("/openapi/updateuserpassword", openapi.Updateuserpassword)
	e.POST("/openapi/updateuserpassword", openapi.Updateuserpassword)
	e.GET("/openapi/updateUserPassword", openapi.UpdateuserpasswordV2)
	e.POST("/openapi/updateUserPassword", openapi.UpdateuserpasswordV2)
	e.GET("/openapi/setUserInfoSelfDefined", openapi.SetUserInfoSelfDefined)
	e.POST("/openapi/setUserInfoSelfDefined", openapi.SetUserInfoSelfDefined)

	e.GET("/openapi/oauth20/getaccesstoken", oauth20.GetAccessToken)
	e.POST("/openapi/oauth20/getaccesstoken", oauth20.GetAccessToken)
	e.GET("/openapi/oauth20/getuserinfo", oauth20.GetUserInfo)
	e.POST("/openapi/oauth20/getuserinfo", oauth20.GetUserInfo)
	e.GET("/openapi/oauth20/getqrcodecontent", oauth20.GetQRCodeContent)
	e.POST("/openapi/oauth20/getqrcodecontent", oauth20.GetQRCodeContent)
	e.GET("/openapi/oauth20/getqrcoderesultcode", oauth20.GetQRCodeResultCode)
	e.POST("/openapi/oauth20/getqrcoderesultcode", oauth20.GetQRCodeResultCode)
	e.GET("/openapi/oauth20/setqrcodeuserinfo", oauth20.SetQRCodeUserinfo)
	e.POST("/openapi/oauth20/setqrcodeuserinfo", oauth20.SetQRCodeUserinfo)

	e.GET("/openapi/rbac/createRole", rbac.CreateRole)
	e.POST("/openapi/rbac/createRole", rbac.CreateRole)
	e.GET("/openapi/rbac/addRoleMap", rbac.AddRoleMap)
	e.POST("/openapi/rbac/addRoleMap", rbac.AddRoleMap)
	e.GET("/openapi/rbac/createPrivilege", rbac.CreatePrivilege)
	e.POST("/openapi/rbac/createPrivilege", rbac.CreatePrivilege)
	e.GET("/openapi/rbac/getRoleList", rbac.GetRoleList)
	e.POST("/openapi/rbac/getRoleList", rbac.GetRoleList)
	e.GET("/openapi/rbac/getPRoleDetail", rbac.GetRoleDetail)
	e.POST("/openapi/rbac/getRoleDetail", rbac.GetRoleDetail)
	e.GET("/openapi/rbac/getPrivilegeList", rbac.GetPrivilegeList)
	e.POST("/openapi/rbac/getPrivilegeList", rbac.GetPrivilegeList)
	e.GET("/openapi/rbac/getPrivilegeDetail", rbac.GetPrivilegeDetail)
	e.POST("/openapi/rbac/getPrivilegeDetail", rbac.GetPrivilegeDetail)
	e.GET("/openapi/rbac/checkPermission", rbac.CheckPermission)
	e.POST("/openapi/rbac/checkPermission", rbac.CheckPermission)
	e.GET("/innerapi/rbac/checkPermission", rbac.CheckPermission)
	e.POST("/innerapi/rbac/checkPermission", rbac.CheckPermission)

	e.GET("/openapi/rbac/createTemplateRole", rbac.CreateTemplateRole)
	e.POST("/openapi/rbac/createTemplateRole", rbac.CreateTemplateRole)
	e.GET("/openapi/rbac/getTemplateRoleList", rbac.GetTemplateRoleList)
	e.POST("/openapi/rbac/getTemplateRoleList", rbac.GetTemplateRoleList)
	e.GET("/openapi/rbac/getTemplateRoleDetail", rbac.GetTemplateRoleDetail)
	e.POST("/openapi/rbac/getTemplateRoleDetail", rbac.GetTemplateRoleDetail)
	e.GET("/openapi/rbac/updateTemplateRole", rbac.UpdateTemplateRole)
	e.POST("/openapi/rbac/updateTemplateRole", rbac.UpdateTemplateRole)
	e.GET("/openapi/rbac/getTemplateRoleList", rbac.GetTemplateRoleList)
	e.POST("/openapi/rbac/getTemplateRoleList", rbac.GetTemplateRoleList)
	e.GET("/innerapi/rbac/generateTemplateRoleTemplate", rbac.GenerateTemplateRoleTemplate)
	e.POST("/innerapi/rbac/generateTemplateRoleTemplate", rbac.GenerateTemplateRoleTemplate)
	e.GET("/innerapi/rbac/createTemplateRole", rbac.CreateTemplateRole)
	e.POST("/innerapi/rbac/createTemplateRole", rbac.CreateTemplateRole)

	e.GET("/openapi/corporation/createCorporation", corporation.CreateCorporation)
	e.POST("/openapi/corporation/createCorporation", corporation.CreateCorporation)
	e.GET("/openapi/corporation/updateCorporation", corporation.UpdateCorporation)
	e.POST("/openapi/corporation/updateCorporation", corporation.UpdateCorporation)
	e.GET("/openapi/corporation/getCorporationDetail", corporation.GetCorporationDetail)
	e.POST("/openapi/corporation/getCorporationDetail", corporation.GetCorporationDetail)
	e.GET("/openapi/corporation/getCorporationAllList", corporation.GetCorporationAllList)
	e.POST("/openapi/corporation/getCorporationAllList", corporation.GetCorporationAllList)
	e.GET("/openapi/corporation/createCorporationUser", corporation.CreateCorporationUser)
	e.POST("/openapi/corporation/createCorporationUser", corporation.CreateCorporationUser)
	e.GET("/openapi/corporation/getUserCorporationList", corporation.GetUserCorporationList)
	e.POST("/openapi/corporation/getUserCorporationList", corporation.GetUserCorporationList)
	e.GET("/innerapi/corporation/getUserDefaultCorporation", corporation.GetUserDefaultCorporation)
	e.POST("/innerapi/corporation/getUserDefaultCorporation", corporation.GetUserDefaultCorporation)
	e.GET("/openapi/corporation/updateCorporationUser", corporation.UpdateCorporationUser)
	e.POST("/openapi/corporation/updateCorporationUser", corporation.UpdateCorporationUser)
	e.GET("/openapi/corporation/getCorporationUserList", corporation.GetCorporationUserList)
	e.POST("/openapi/corporation/getCorporationUserList", corporation.GetCorporationUserList)
	e.GET("/openapi/corporation/getCorporationUserAllList", corporation.GetCorporationUserAllList)
	e.POST("/openapi/corporation/getCorporationUserAllList", corporation.GetCorporationUserAllList)
	e.GET("/openapi/corporation/getCorporationUserDetail", corporation.GetCorporationUserDetail)
	e.POST("/openapi/corporation/getCorporationUserDetail", corporation.GetCorporationUserDetail)

	e.GET("/innerapi/getuserinfo", innerapi.GetUserInfo)
	e.POST("/innerapi/getuserinfo", innerapi.GetUserInfo)
	e.GET("/innerapi/getUserInfo", innerapi.GetUserInfoV2)
	e.POST("/innerapi/getUserInfo", innerapi.GetUserInfoV2)

	e.Logger.Fatal(e.Start(":" + config.Config().Port))
}
