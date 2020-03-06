/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package oauth20

import (
	"github.com/qoobing/userd/src/api/oauth20/getuserinfo"
	"github.com/qoobing/userd/src/api/oauth20/getuserinfoaccesstoken"
	"github.com/qoobing/userd/src/api/oauth20qrcode/getqrcodecontent"
	"github.com/qoobing/userd/src/api/oauth20qrcode/getqrcoderesultcode"
	"github.com/qoobing/userd/src/api/oauth20qrcode/setqrcodeuserinfo"
)

var (
	GetUserInfo         = getuserinfo.Main
	GetAccessToken      = getuserinfoaccesstoken.Main
	GetQRCodeContent    = getqrcodecontent.Main
	GetQRCodeResultCode = getqrcoderesultcode.Main
	SetQRCodeUserinfo   = setqrcodeuserinfo.Main
)
