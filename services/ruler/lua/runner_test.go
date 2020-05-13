package lua_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/shibukawa/configdir"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler/lua"
	"github.com/wealdtech/walletd/services/storage/mem"
)

func TestStorage(t *testing.T) {

	configDirs := configdir.New("wealdtech", "walletd")
	storageFile := filepath.Join(configDirs.QueryFolders(configdir.Global)[0].Path, "scripts", "count.lua")
	fmt.Printf("storageFile is %s\n", storageFile)
	defer os.Remove(storageFile)
	err := ioutil.WriteFile(storageFile, []byte(`function approve(request, storage, messages)
  if storage.anumber == nil then
    storage.anumber = 1
    storage.abool = true
    storage.astring = "echo"
  end

  table.insert(messages, string.format("The number: %d", storage.anumber))
  table.insert(messages, string.format("The bool: %t", storage.abool))
  table.insert(messages, string.format("The string: %s", storage.astring))

  storage.anumber = storage.anumber + 1
  storage.abool = not storage.abool
  storage.astring = storage.astring .. " echo"

  return "Approved"
end`), 0644)
	require.NoError(t, err)

	locker, err := locker.New()
	require.NoError(t, err)
	store, err := mem.New()
	require.NoError(t, err)

	ruleDefs := make([]*core.RuleDefinition, 0)
	ruleDefs = append(ruleDefs, &core.RuleDefinition{
		Name:    "test",
		Request: "sign",
		Account: ".*",
		Script:  "count.lua",
	})
	rules, err := core.InitRules(context.Background(), ruleDefs)
	require.NoError(t, err)

	ruler, err := lua.New(locker, store, rules)
	require.NoError(t, err)

	result := ruler.RunRules(context.Background(), "sign", "Test wallet", "Test account", []byte{}, nil)
	fmt.Printf("Result is %v\n", result)

	result2 := ruler.RunRules(context.Background(), "sign", "Test wallet", "Test account", []byte{}, nil)
	fmt.Printf("Result 2 is %v\n", result2)
}
