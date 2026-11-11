package runtimesvc

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/platform"
)

type recordedEvent struct {
	channel string
	data    any
}

type recordingEvents struct {
	events []recordedEvent
}

func (r *recordingEvents) Emit(channel string, data any) {
	r.events = append(r.events, recordedEvent{channel: channel, data: data})
}

var expectedMethods = []string{
	"RunGame", "ListPrefixes", "CreatePrefix", "GetPrefixBaseDir", "GetPrefixCreatedAt",
	"GetPrefixStats", "RemovePrefix", "SavePrefixConfig", "LoadPrefixConfig",
	"LoadPrefixConfigWithProton", "RunPrefixTool", "ScanProtonVersions", "GetProtonVariants",
	"GetProtonReleases", "InstallProtonVersion", "GetLsfgProfileForGame", "DetectLosslessDll",
	"SaveLsfgProfile", "DisableLsfgProfile", "RemoveProfile", "InstallLsfg", "UninstallLsfg",
	"GetRunningSessions", "KillSession",
}

func TestRegisterAddsExactlyTheRuntimeMethods(t *testing.T) {
	router := api.NewRouter()
	Register(router, &recordingEvents{})

	got := router.Methods()
	sort.Strings(got)
	want := append([]string(nil), expectedMethods...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("methods =\n%v\nwant\n%v", got, want)
	}
}

func TestUnsupportedPlatformListsAreEmpty(t *testing.T) {
	handler := listOrEmpty(false, plain(func() (any, error) {
		t.Fatal("the real handler must not run when the feature is off")
		return nil, nil
	}))

	result, err := handler(context.Background(), nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	data, _ := json.Marshal(result)
	if string(data) != "[]" {
		t.Errorf("list = %s, want []", data)
	}
}

func TestUnsupportedPlatformMethodsReturnClearError(t *testing.T) {
	handler := gated(false, "prefix", plain(func() (any, error) {
		t.Fatal("the real handler must not run when the feature is off")
		return nil, nil
	}))

	_, err := handler(context.Background(), nil)
	want := "prefix not supported on " + platform.OS()
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestDispatchCreatesPrefixFromPayload(t *testing.T) {
	if !platform.Current().Prefix {
		t.Skip("prefixes are not supported on this platform")
	}
	base := isolateConfigHome(t)
	router := api.NewRouter()
	Register(router, &recordingEvents{})

	payload := json.RawMessage(`{"name": "Game One"}`)
	if _, err := router.Dispatch(context.Background(), "CreatePrefix", payload); err != nil {
		t.Fatalf("CreatePrefix: %v", err)
	}
	if !config.PathExists(filepath.Join(base, "Game One")) {
		t.Error("CreatePrefix did not create the prefix named in the payload")
	}
}

func TestProgressEventUsesUIFieldNames(t *testing.T) {
	events := &recordingEvents{}
	emit := progressEmitter(events, "install-proton-progress")
	emit(50, "Extracting Proton package...")

	if len(events.events) != 1 || events.events[0].channel != "install-proton-progress" {
		t.Fatalf("events = %+v, want one install-proton-progress event", events.events)
	}
	data, err := json.Marshal(events.events[0].data)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"percent":50,"message":"Extracting Proton package..."}` {
		t.Errorf("event JSON = %s", data)
	}
}

func TestResponseJSONShapes(t *testing.T) {
	size := int64(8)
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"prefix stats", PrefixStats{Name: "Game"}, `{"name":"Game","createdAt":null,"sizeBytes":null}`},
		{"prefix stats with size", PrefixStats{Name: "Game", SizeBytes: &size}, `{"name":"Game","createdAt":null,"sizeBytes":8}`},
		{"variant", knownVariants[0], `{"ID":"ge-proton","Name":"GE-Proton (GloriousEggroll)","Description":"The most popular custom Proton build. Includes many game fixes and codec patches.","RepoOwner":"GloriousEggroll","RepoName":"proton-ge-custom"}`},
	}
	for _, test := range tests {
		data, err := json.Marshal(test.value)
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if string(data) != test.want {
			t.Errorf("%s JSON = %s, want %s", test.name, data, test.want)
		}
	}
}
