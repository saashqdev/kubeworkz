/*
Copyright 2024 Kubeworkz Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errcode

var (
	MissingParamUserName             = New(missingParam, "name")
	MissingParamPassword             = New(missingParam, "password")
	InvalidParameterPassword         = New(invalidParamValue, "password")
	InvalidParameterPhone            = New(invalidParamValue, "phone")
	InvalidParameterEmail            = New(invalidParamValue, "email")
	MissingParamFile                 = New(missingParam, "file")
	MissingParamNameOrPwdOrLoginType = New(missingParam, "name or password or login type")

	AuthenticateError = New(authenticateError)
	UserNotExist      = New(userNotExist)
	InvalidToken      = New(invalidToken)
	ForbiddenErr      = New(forbidden)
	LdapConnectError  = New(ldapConnectError)
	PasswordWrong     = New(passwordWrong)
	UserIsDisabled    = New(userIsDisabled)
)

func UserNameDuplicated(name string) *ErrorInfo {
	return New(paramNotUnique, "name", name)
}
