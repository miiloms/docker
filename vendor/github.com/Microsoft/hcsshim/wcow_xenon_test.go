package hcsshim

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// --------------------------------
//    W C O W    X E N O N   V 1
// --------------------------------

// A v1 WCOW Xenon with a single base layer
func TestV1XenonWCOW(t *testing.T) {
	t.Skip("for now")
	tempDir := createWCOWTempDirWithSandbox(t)
	defer os.RemoveAll(tempDir)

	layers := layersNanoserver
	uvmImagePath, err := LocateWCOWUVMFolderFromLayerFolders(layers)
	if err != nil {
		t.Fatalf("LocateWCOWUVMFolderFromLayerFolders failed %s", err)
	}
	options := make(map[string]string)
	options[HCSOPTION_SCHEMA_VERSION] = SchemaV10().String()
	c, err := CreateContainerEx(&CreateOptions{
		Id:      "TestV1XenonWCOW",
		Options: options,
		Spec: &specs.Spec{
			Windows: &specs.Windows{
				LayerFolders: append(layers, tempDir),
				HyperV:       &specs.WindowsHyperV{UtilityVMPath: filepath.Join(uvmImagePath, "UtilityVM")},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed create: %s", err)
	}
	startContainer(t, c)
	runCommand(t, c, "cmd /s /c echo Hello", `c:\`, "Hello")
	stopContainer(t, c)
}

// A v1 WCOW Xenon with a single base layer but let HCSShim find the utility VM path
func TestV1XenonWCOWNoUVMPath(t *testing.T) {
	t.Skip("for now")
	tempDir := createWCOWTempDirWithSandbox(t)
	defer os.RemoveAll(tempDir)

	options := make(map[string]string)
	options[HCSOPTION_SCHEMA_VERSION] = SchemaV10().String()
	c, err := CreateContainerEx(&CreateOptions{
		Id:      "TestV1XenonWCOWNoUVMPath",
		Owner:   "unit-test",
		Options: options,
		Spec: &specs.Spec{
			Windows: &specs.Windows{
				LayerFolders: append(layersNanoserver, tempDir),
				HyperV:       &specs.WindowsHyperV{},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed create: %s", err)
	}
	startContainer(t, c)
	runCommand(t, c, "cmd /s /c echo Hello", `c:\`, "Hello")
	stopContainer(t, c)
}

// A v1 WCOW Xenon with multiple layers letting HCSShim find the utilityVM Path
func TestV1XenonMultipleBaseLayersNoUVMPath(t *testing.T) {
	t.Skip("for now")
	tempDir := createWCOWTempDirWithSandbox(t)
	defer os.RemoveAll(tempDir)

	layers := layersBusybox
	options := make(map[string]string)
	options[HCSOPTION_SCHEMA_VERSION] = SchemaV10().String()
	c, err := CreateContainerEx(&CreateOptions{
		Id:      "TestV1XenonWCOW",
		Options: options,
		Spec: &specs.Spec{
			Windows: &specs.Windows{
				LayerFolders: append(layers, tempDir),
				HyperV:       &specs.WindowsHyperV{},
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed create: %s", err)
	}
	startContainer(t, c)
	runCommand(t, c, "cmd /s /c echo Hello", `c:\`, "Hello")
	stopContainer(t, c)
}

// --------------------------------
//    W C O W    X E N O N   V 2
// --------------------------------

// A single WCOW xenon. Note in this test, neither the UVM or the
// containers are supplied IDs - they will be autogenerated for us.
// This is the minimum set of parameters needed to create a V2 WCOW xenon.
func TestV2XenonWCOW(t *testing.T) {
	t.Skip("Skipping for now")
	uvm, uvmScratchDir := createv2WCOWUVM(t, layersNanoserver, "", nil)
	defer os.RemoveAll(uvmScratchDir)
	defer uvm.Terminate()
	if err := uvm.Start(); err != nil {
		t.Fatalf("Failed start utility VM: %s", err)
	}

	// Create the container hosted inside the utility VM
	containerScratchDir := createWCOWTempDirWithSandbox(t)
	defer os.RemoveAll(containerScratchDir)
	options := make(map[string]string)
	layerFolders := append(layersNanoserver, containerScratchDir)
	hostedContainer, err := CreateContainerEx(&CreateOptions{
		HostingSystem: uvm,
		Options:       options,
		Spec:          &specs.Spec{Windows: &specs.Windows{LayerFolders: layerFolders}},
	})
	if err != nil {
		t.Fatalf("CreateContainerEx failed: %s", err)
	}
	defer UnmountContainerLayers(layerFolders, uvm, UnmountOperationAll)

	// Start/stop the container
	startContainer(t, hostedContainer)
	runCommand(t, hostedContainer, "cmd /s /c echo TestV2XenonWCOW", `c:\`, "TestV2XenonWCOW")
	stopContainer(t, hostedContainer)
	hostedContainer.Terminate()
}

// TODO: Have a similar test where the UVM scratch folder does not exist.
// A single WCOW xenon but where the container sandbox folder is not pre-created by the client
func TestV2XenonWCOWContainerSandboxFolderDoesNotExist(t *testing.T) {
	t.Skip("Skipping for now")
	uvm, uvmScratchDir := createv2WCOWUVM(t, layersNanoserver, "TestV2XenonWCOWContainerSandboxFolderDoesNotExist_UVM", nil)
	defer os.RemoveAll(uvmScratchDir)
	defer uvm.Terminate()
	if err := uvm.Start(); err != nil {
		t.Fatalf("Failed start utility VM: %s", err)
	}

	// This is the important bit for this test. It's deleted here. We call the helper only to allocate a temporary directory
	containerScratchDir := createWCOWTempDirWithSandbox(t)
	os.RemoveAll(containerScratchDir)

	layerFolders := append(layersBusybox, containerScratchDir)
	hostedContainer, err := CreateContainerEx(&CreateOptions{
		Id:            "container",
		HostingSystem: uvm,
		Spec:          &specs.Spec{Windows: &specs.Windows{LayerFolders: layerFolders}},
	})
	if err != nil {
		t.Fatalf("CreateContainerEx failed: %s", err)
	}
	defer UnmountContainerLayers(layerFolders, uvm, UnmountOperationAll)

	// Start/stop the container
	startContainer(t, hostedContainer)
	runCommand(t, hostedContainer, "cmd /s /c echo TestV2XenonWCOW", `c:\`, "TestV2XenonWCOW")
	stopContainer(t, hostedContainer)
	hostedContainer.Terminate()
}

// TODO What about mount. Test with the client doing the mount.
// TODO Test as above, but where sandbox for UVM is entirely created by a client to show how it's done.

// Two v2 WCOW containers in the same UVM, each with a single base layer
func TestV2XenonWCOWTwoContainers(t *testing.T) {
	t.Skip("Skipping for now")
	uvm, uvmScratchDir := createv2WCOWUVM(t, layersNanoserver, "TestV2XenonWCOWTwoContainers_UVM", nil)
	defer os.RemoveAll(uvmScratchDir)
	defer uvm.Terminate()
	if err := uvm.Start(); err != nil {
		t.Fatalf("Failed start utility VM: %s", err)
	}

	// First hosted container
	firstContainerScratchDir := createWCOWTempDirWithSandbox(t)
	defer os.RemoveAll(firstContainerScratchDir)
	options := make(map[string]string)
	options[HCSOPTION_SCHEMA_VERSION] = SchemaV20().String()
	firstLayerFolders := append(layersNanoserver, firstContainerScratchDir)
	firstHostedContainer, err := CreateContainerEx(&CreateOptions{
		Id:            "FirstContainer",
		HostingSystem: uvm,
		Options:       options,
		Spec:          &specs.Spec{Windows: &specs.Windows{LayerFolders: firstLayerFolders}},
	})
	if err != nil {
		t.Fatalf("CreateContainerEx failed: %s", err)
	}
	defer UnmountContainerLayers(firstLayerFolders, uvm, UnmountOperationAll)

	// Second hosted container
	secondContainerScratchDir := createWCOWTempDirWithSandbox(t)
	defer os.RemoveAll(firstContainerScratchDir)
	secondLayerFolders := append(layersNanoserver, secondContainerScratchDir)
	secondHostedContainer, err := CreateContainerEx(&CreateOptions{
		Id:            "SecondContainer",
		HostingSystem: uvm,
		Options:       options,
		Spec:          &specs.Spec{Windows: &specs.Windows{LayerFolders: secondLayerFolders}},
	})
	if err != nil {
		t.Fatalf("CreateContainerEx failed: %s", err)
	}
	defer UnmountContainerLayers(secondLayerFolders, uvm, UnmountOperationAll)

	startContainer(t, firstHostedContainer)
	runCommand(t, firstHostedContainer, "cmd /s /c echo FirstContainer", `c:\`, "FirstContainer")
	startContainer(t, secondHostedContainer)
	runCommand(t, secondHostedContainer, "cmd /s /c echo SecondContainer", `c:\`, "SecondContainer")
	stopContainer(t, firstHostedContainer)
	stopContainer(t, secondHostedContainer)
	firstHostedContainer.Terminate()
	secondHostedContainer.Terminate()
}

//// This verifies the container storage is unmounted correctly so that a second
//// container can be started from the same storage.
//func TestV2XenonWCOWWithRemount(t *testing.T) {
////	t.Skip("Skipping for now")
//	uvmID := "Testv2XenonWCOWWithRestart_UVM"
//	uvmScratchDir, err := ioutil.TempDir("", "uvmScratch")
//	if err != nil {
//		t.Fatalf("Failed create temporary directory: %s", err)
//	}
//	if err := CreateWCOWSandbox(layersNanoserver[0], uvmScratchDir, uvmID); err != nil {
//		t.Fatalf("Failed create Windows UVM Sandbox: %s", err)
//	}
//	defer os.RemoveAll(uvmScratchDir)

//	uvm, err := CreateContainerEx(&CreateOptions{
//		Id:              uvmID,
//		Owner:           "unit-test",
//		SchemaVersion:   SchemaV20(),
//		IsHostingSystem: true,
//		Spec: &specs.Spec{
//			Windows: &specs.Windows{
//				LayerFolders: []string{uvmScratchDir},
//				HyperV:       &specs.WindowsHyperV{UtilityVMPath: filepath.Join(layersNanoserver[0], `UtilityVM\Files`)},
//			},
//		},
//	})
//	if err != nil {
//		t.Fatalf("Failed create UVM: %s", err)
//	}
//	defer uvm.Terminate()
//	if err := uvm.Start(); err != nil {
//		t.Fatalf("Failed start utility VM: %s", err)
//	}

//	// Mount the containers storage in the utility VM
//	containerScratchDir := createWCOWTempDirWithSandbox(t)
//	layerFolders := append(layersNanoserver, containerScratchDir)
//	cls, err := Mount(layerFolders, uvm, SchemaV20())
//	if err != nil {
//		t.Fatalf("failed to mount container storage: %s", err)
//	}
//	combinedLayers := cls.(CombinedLayersV2)
//	mountedLayers := &ContainersResourcesStorageV2{
//		Layers: combinedLayers.Layers,
//		Path:   combinedLayers.ContainerRootPath,
//	}
//	defer func() {
//		if err := Unmount(layerFolders, uvm, SchemaV20(), UnmountOperationAll); err != nil {
//			t.Fatalf("failed to unmount container storage: %s", err)
//		}
//	}()

//	// Create the first container
//	defer os.RemoveAll(containerScratchDir)
//	xenon, err := CreateContainerEx(&CreateOptions{
//		Id:            "container",
//		Owner:         "unit-test",
//		HostingSystem: uvm,
//		SchemaVersion: SchemaV20(),
//		Spec:          &specs.Spec{Windows: &specs.Windows{}}, // No layerfolders as we mounted them ourself.
//	})
//	if err != nil {
//		t.Fatalf("CreateContainerEx failed: %s", err)
//	}

//	// Start/stop the first container
//	startContainer(t, xenon)
//	runCommand(t, xenon, "cmd /s /c echo TestV2XenonWCOWFirstStart", `c:\`, "TestV2XenonWCOWFirstStart")
//	stopContainer(t, xenon)
//	xenon.Terminate()

//	// Now unmount and remount to exactly the same places
//	if err := Unmount(layerFolders, uvm, SchemaV20(), UnmountOperationAll); err != nil {
//		t.Fatalf("failed to unmount container storage: %s", err)
//	}
//	if _, err = Mount(layerFolders, uvm, SchemaV20()); err != nil {
//		t.Fatalf("failed to mount container storage: %s", err)
//	}

//	// Create an identical second container and verify it works too.
//	xenon2, err := CreateContainerEx(&CreateOptions{
//		Id:            "container",
//		Owner:         "unit-test",
//		HostingSystem: uvm,
//		SchemaVersion: SchemaV20(),
//		Spec:          &specs.Spec{Windows: &specs.Windows{LayerFolders: layerFolders}},
//		MountedLayers: mountedLayers,
//	})
//	if err != nil {
//		t.Fatalf("CreateContainerEx failed: %s", err)
//	}
//	startContainer(t, xenon2)
//	runCommand(t, xenon2, "cmd /s /c echo TestV2XenonWCOWAfterRemount", `c:\`, "TestV2XenonWCOWAfterRemount")
//	stopContainer(t, xenon2)
//	xenon2.Terminate()
//}

// Lots of v2 WCOW containers in the same UVM, each with a single base layer. Containers aren't
// actually started, but it stresses the SCSI controller hot-add logic.
func TestV2XenonWCOWCreateLots(t *testing.T) {
	t.Skip("Skipping for now")
	uvm, uvmScratchDir := createv2WCOWUVM(t, layersNanoserver, "TestV2XenonWCOWTwoContainers_UVM", nil)
	defer os.RemoveAll(uvmScratchDir)
	defer uvm.Terminate()
	if err := uvm.Start(); err != nil {
		t.Fatalf("Failed start utility VM: %s", err)
	}

	for i := 0; i < 64; i++ {
		containerScratchDir := createWCOWTempDirWithSandbox(t)
		defer os.RemoveAll(containerScratchDir)
		options := make(map[string]string)
		options[HCSOPTION_SCHEMA_VERSION] = SchemaV20().String()
		layerFolders := append(layersNanoserver, containerScratchDir)
		hostedContainer, err := CreateContainerEx(&CreateOptions{
			Id:            fmt.Sprintf("container%d", i),
			HostingSystem: uvm,
			Options:       options,
			Spec:          &specs.Spec{Windows: &specs.Windows{LayerFolders: layerFolders}},
		})
		if err != nil {
			t.Fatalf("CreateContainerEx failed: %s", err)
		}
		defer hostedContainer.Terminate()
		defer UnmountContainerLayers(layerFolders, uvm, UnmountOperationAll)
	}

	// TODO: Push it over 64 now and will get a failure.
}

// Helper for the v2 Xenon tests to create a utility VM. Returns the container
// object; folder used as its scratch
func createv2WCOWUVM(t *testing.T, uvmLayers []string, uvmId string, resources *specs.WindowsResources) (Container, string) {
	uvmScratchDir := createTempDir(t)
	spec := &specs.Spec{Windows: &specs.Windows{LayerFolders: append(uvmLayers, uvmScratchDir)}}
	if resources != nil {
		spec.Windows.Resources = resources
	}
	options := make(map[string]string)
	options[HCSOPTION_SCHEMA_VERSION] = SchemaV20().String()
	options[HCSOPTION_IS_UTILITY_VM] = "yes"

	createOptions := &CreateOptions{
		Options: options,
		Spec:    spec,
	}
	if uvmId != "" {
		createOptions.Id = uvmId
	}

	uvm, err := CreateContainerEx(createOptions)
	if err != nil {
		t.Fatalf("Failed create UVM: %s", err)
	}
	return uvm, uvmScratchDir
}

// TestV2XenonWCOWMultiLayer creates a V2 Xenon having multiple image layers
func TestV2XenonWCOWMultiLayer(t *testing.T) {
	//t.Skip("for now")

	uvmMemory := uint64(1 * 1024 * 1024 * 1024)
	uvmCPUCount := uint64(2)
	resources := &specs.WindowsResources{
		Memory: &specs.WindowsMemoryResources{
			Limit: &uvmMemory,
		},
		CPU: &specs.WindowsCPUResources{
			Count: &uvmCPUCount,
		},
	}
	uvm, uvmScratchDir := createv2WCOWUVM(t, layersNanoserver, "TestV2XenonWCOWMultiLayer_UVM", resources)
	defer os.RemoveAll(uvmScratchDir)
	defer uvm.Terminate()
	if err := uvm.Start(); err != nil {
		t.Fatalf("Failed start utility VM: %s", err)
	}

	// Create a sandbox for the hosted container
	containerScratchDir := createWCOWTempDirWithSandbox(t)
	defer os.RemoveAll(containerScratchDir)

	// Create the container
	options := make(map[string]string)
	options[HCSOPTION_SCHEMA_VERSION] = SchemaV20().String()
	containerLayers := append(layersBusybox, containerScratchDir)
	xenon, err := CreateContainerEx(&CreateOptions{
		Id:            "container",
		HostingSystem: uvm,
		Options:       options,
		Spec:          &specs.Spec{Windows: &specs.Windows{LayerFolders: containerLayers}},
	})
	if err != nil {
		t.Fatalf("CreateContainerEx failed: %s", err)
	}

	// Start/stop the container
	startContainer(t, xenon)
	runCommand(t, xenon, "echo Container", `c:\`, "Container")
	stopContainer(t, xenon)
	xenon.Terminate()
	if err := UnmountContainerLayers(containerLayers, uvm, UnmountOperationAll); err != nil {
		t.Fatalf("unmount failed: %s", err)
	}

}
