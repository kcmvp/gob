package scaffolds

import (
	"fmt"
	"log"

	"github.com/kcmvp/gob/boot"
)

// saveHistory.
var saveHistory boot.Action = func(session *boot.Session, project boot.Project, command boot.Command) error {
	var err error
	hash := "latest"
	if git := project.Git(); git != nil {
		wt, err := git.Worktree()
		if err != nil {
			return fmt.Errorf("Failed to save build history report:%w", err)
		}
		st, err := wt.Status()
		if err != nil {
			return fmt.Errorf("Failed to save build history report:%w", err)
		}
		// get hash only when worktree is clean
		if st.IsClean() {
			ref, err := git.Head()
			if err != nil {
				return fmt.Errorf("Failed to save build history report:%w", err)
			}
			hash = ref.Hash().String()[0:8]
		}
	}
	log.Println(hash)
	// 1: save - 1
	// 1.1 key: {current}-{parent}-{0} : in order to quickly get the chain
	// 1.2: stale data is need to take into consideration (latest-xxx) but the parent is not the correct parent
	//     save only latest x records, delete the previous records after successfully save
	// 1.3:  fix : In this case for latest we need always check(delete and insert) when save data.
	// 1.4: re-run : don't save data when {c-hash}-{p-hash} does not change: one-exception: c-hash is 'latest'
	// 1.5: key will changed to {current}-{parent}-1

	// 2: query - need to scan the workspace first then build out the tree and qury db to show the result

	// 3: delete - in order to quickly delete result, need to record the commit hash creation time (might be use
	// another bulk to save the result).

	return err
}
