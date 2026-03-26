package verify

import (
	"regexp"
	"unicode"
)

func VerifyPasswordFormat(password string) bool {
	var hasNumber, hasLetter, hasUpperCase, hasLowercase, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			hasNumber = true //是否包含数字
		case unicode.IsLetter(c):
			hasLetter = true //是否包含字母
			if unicode.IsUpper(c) {
				hasUpperCase = true //是否包含大写字母
			}
			if unicode.IsLower(c) {
				hasLowercase = true //是否包含小写字母
			}
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true //是否包含特殊符号
		}

	}
	//不检查特殊字符
	hasSpecial = true

	//不检查大写
	hasUpperCase = true

	//不检查小写
	hasLowercase = true

	return hasNumber && hasLetter && hasUpperCase && hasLowercase && hasSpecial && len(password) > 7 && len(password) < 21
}

func VerifyPayPassFormat(password string) bool {
	regular := "^[0-9]{6,6}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(password)

}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func VerifyMobileFormat(mobileNum string) bool {
	regular := "^((13[0-9])|(14[4-9])|(15[0-3,5-9])|(16[0-9])|(17[0-8])|(18[0-9])|(19[0-9]))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

// 匹配中国手机号码
func MatchChinaMobile(mobileNo string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNo)
}

// 匹配邮箱
func MatchEmail(email string) bool {
	//pattern := `\w+([-+.]\w+)@\w+([-.]\w+).\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z].){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 匹配身份证号码
func MatchIdCard(idCard string) bool {
	regular := "^[1-9]\\d{5}(18|19|([23]\\d))\\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\\d{3}[0-9Xx]$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(idCard)
}

// 匹配银行卡
func MatchBankCard(bankCard string) bool {
	regular := "^([1-9]{1})(\\d{14}|\\d{18})$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(bankCard)
}
