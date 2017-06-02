package httpd

import (
	"fmt"

	"github.com/lodastack/registry/config"

	"github.com/go-ldap/ldap"
)

// LDAPAuth for auth user
func LDAPAuth(username string, password string) error {
	if password == "" || username == "" {
		return fmt.Errorf("need username or password")
	}

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s", config.C.LDAPConf.Server))
	if err != nil {
		return err
	}
	defer l.Close()

	// First bind with a read only user
	err = l.Bind(config.C.LDAPConf.Binddn, config.C.LDAPConf.Password)
	if err != nil {
		return err
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		config.C.LDAPConf.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("((%s=%s))", config.C.LDAPConf.UID, username),
		[]string{""},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return err
	}

	if len(sr.Entries) != 1 {
		return fmt.Errorf("User does not exist or too many entries returned: %d", len(sr.Entries))
	}

	userdn := sr.Entries[0].DN

	// Bind as the user to verify their password
	err = l.Bind(userdn, password)
	if err != nil {
		return err
	}
	return nil
}

func LDAPUserExist(username string) bool {
	if username == "" {
		return false
	}

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s", config.C.LDAPConf.Server))
	if err != nil {
		return false
	}
	defer l.Close()

	// First bind with a read only user
	if err = l.Bind(config.C.LDAPConf.Binddn, config.C.LDAPConf.Password); err != nil {
		return false
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		config.C.LDAPConf.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("((%s=%s))", config.C.LDAPConf.UID, username),
		[]string{""},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return false
	}
	if len(sr.Entries) != 1 {
		return false
	}
	return true
}
