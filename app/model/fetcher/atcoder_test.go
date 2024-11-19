package fetcher

import (
	"testing"
	"time"

	"github.com/YourSuzumiya/ACMBot/app/model/errs"
)

var (
	existedHandles   = []string{"tourist", "rng_58", "chokudai", "gori", "petr", "jiangly"}
	unexistedHandles = []string{"dongxuelian", "daijiangqi", "aminuosi"}
)

func TestExistedAtcoderUser(t *testing.T) {
	for _, handle := range existedHandles {
		user, err := FetchAtcoderUser(handle)
		if err != nil {
			t.Errorf("Failed to fetch Atcoder user: %+v", err)
		} else {
			t.Logf("User: %+v", user)
		}
	}
}

func TestUnExistedAtcoderUser(t *testing.T) {
	for _, handle := range unexistedHandles {
		if user, err := FetchAtcoderUser(handle); err != errs.ErrHandleNotFound {
			t.Error("Fetching unexisted user should return ErrHandleNotFound, But got: ", err)
            if err == nil {
                t.Logf("User %+v", user)
            }
		}
	}
}

func TestAtcoderUserSubmissionsFromEpoch(t *testing.T) {
	for _, handle := range existedHandles {
		subs, err := FetchAtcoderUserSubmissionList(handle, time.Unix(0, 0).Unix())
		if err != nil {
			t.Error(err)
		} else if len(*subs) == 0 {
			t.Logf("Existed user %v doesn't has any submission", handle)
		}
	}

	for _, handle := range unexistedHandles {
		subs, err := FetchAtcoderUserSubmissionList(handle, time.Unix(0, 0).Unix())
		if err != nil {
			t.Error(err)
		} else if len(*subs) != 0 {
			t.Errorf("Unexisted user %v submission length should 0 from epoch!", handle)
		}
	}
}

func TestAtcoderUserSubmissionsFromNow(t *testing.T) {
	for _, handle := range existedHandles {
		subs, err := FetchAtcoderUserSubmissionList(handle, time.Now().Unix())
		if err != nil {
			t.Error(err)
		} else if len(*subs) != 0 {
			t.Error("User submission length should 0 from now!")
		}
	}

	for _, handle := range unexistedHandles {
		subs, err := FetchAtcoderUserSubmissionList(handle, time.Now().Unix())
		if err != nil {
			t.Error(err)
		} else if len(*subs) != 0 {
			t.Error("User submission length should 0 from now!")
		}
	}
}

func TestAtcoderContests(t *testing.T) {
	contests, err := FetchAtcoderContestList()
	if err != nil {
		t.Error(err)
	} else if len(*contests) == 0 {
		t.Error("Contest list should not be empty!")
	}
}
