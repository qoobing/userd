package logout

import (
	"github.com/labstack/echo"
	. "github.com/qoobing/userd/src/const"
	"github.com/qoobing/userd/src/model"
	"github.com/qoobing/userd/src/model/t_user"
	. "qoobing.com/utillib.golang/api"
	"qoobing.com/utillib.golang/log"
)

type Input struct {
	UATK string `json:"UATK" validate:"omitempty,min=4"`
}

type Output struct {
	Eno int    `json:"eno"`
	Err string `json:"err"`
}

func Main(cc echo.Context) error {
	//Step 1. init apicontext
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.Redis()
	c.Mysql()

	//Step 2. parameters initial
	var (
		input  Input
		output Output
	)
	if err := c.BindInput(&input); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}

	//删除access_token

	//Step 3. get user info
	//user, err := model.GetLoginInfo(c, c.Redis())
	iuser, err := model.GetUserAccessTokenData(c.Redis(), input.UATK)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

	user := t_user.User{
		F_user_id: iuser.Userid,
	}
	err = model.DelUserAccessTokenData(c.Redis(), user)
	if err != nil {
		log.Debugf(err.Error())
		return c.RESULT_ERROR(ERR_NOT_LOGIN, "not login")
	}

	//Step 4. set output
	output.Eno = 0
	output.Err = "success"

	return c.RESULT(output)
}
