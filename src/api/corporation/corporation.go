/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package corporation

import (
	"github.com/qoobing/userd/src/api/corporation/createCorporation"
	"github.com/qoobing/userd/src/api/corporation/createCorporationUser"
	"github.com/qoobing/userd/src/api/corporation/getCorporationAllList"
	"github.com/qoobing/userd/src/api/corporation/getCorporationDetail"
	"github.com/qoobing/userd/src/api/corporation/getCorporationUserAllList"
	"github.com/qoobing/userd/src/api/corporation/getCorporationUserDetail"
	"github.com/qoobing/userd/src/api/corporation/getCorporationUserList"
	"github.com/qoobing/userd/src/api/corporation/getUserCorporationList"
	"github.com/qoobing/userd/src/api/corporation/getUserDefaultCorporation"
	"github.com/qoobing/userd/src/api/corporation/updateCorporation"
	"github.com/qoobing/userd/src/api/corporation/updateCorporationUser"
)

var (
	CreateCorporation         = createCorporation.Main
	CreateCorporationUser     = createCorporationUser.Main
	UpdateCorporationUser     = updateCorporationUser.Main
	GetCorporationUserList    = getCorporationUserList.Main
	GetCorporationUserDetail  = getCorporationUserDetail.Main
	GetCorporationUserAllList = getCorporationUserAllList.Main
	GetUserCorporationList    = getUserCorporationList.Main
	GetUserDefaultCorporation = getUserDefaultCorporation.Main
	GetCorporationDetail      = getCorporationDetail.Main
	GetCorporationAllList     = getCorporationAllList.Main
	UpdateCorporation         = updateCorporation.Main
)
