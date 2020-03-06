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
package types

type Userinfo struct {
	Name           string `json:"name"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	Loginstate     int    `json:"loginstate"`
	Lastupdatetime string `json:"lastupdatetime"`
	Userid         uint64 `json:"userid"`
}

type VerifyCodeInfo struct {
	Name      string `json:"name"`
	Nametype  int    `json:"nametype"`
	VcodeType int    `json:"vcodeType"`
	Vcode     string `json:"vcode"`
	Vcodekey  string `json:"vcodekey"`
}

type CaptchaInfo struct {
	Sign     string `json:"sign"`
	Vcode    string `json:"vcode"`
	Vcodekey string `json:"vcodekey"`
}
