package clientmanager

import (
	"fmt"
	"sync"

	"github.com/AdamBowers/terraform-provider-pve/v2/internal/diagnostic"
	"github.com/luthermonson/go-proxmox"
)

type ClientManager struct {
	mu      sync.RWMutex
	clients map[string]*proxmox.Client
}

func New() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*proxmox.Client),
	}
}

// Gets a client by name.
// Returns nil if client not found by the provided name.
func (cm *ClientManager) get(name string) *proxmox.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.clients[name]
}

// set stores a client under the name, overwriting any existing client.
func (cm *ClientManager) set(name string, client *proxmox.Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.clients == nil {
		cm.clients = make(map[string]*proxmox.Client)
	}

	cm.clients[name] = client
}

// add stores a client only if no client already exists with the same name.
// t returns true if the client was added and false when the name is in use.
func (cm *ClientManager) add(name string, client *proxmox.Client) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.clients == nil {
		cm.clients = make(map[string]*proxmox.Client)
	}

	if _, exists := cm.clients[name]; exists {
		return false
	}

	cm.clients[name] = client
	return true
}

func (cm *ClientManager) Get(name string, diags *diagnostic.Diagnostics) *proxmox.Client {
	if client := cm.get(name); client != nil {
		return client
	}

	diags.Append(
		diagnostic.Error(
			"Client Not Found",
			fmt.Sprintf("no client found with name %q", name),
			diagnostic.WithoutStacktrace(),
			diagnostic.WithCategory(diagnostic.Category_InvalidParameter),
		),
	)
	return nil
}

func (cm *ClientManager) TryGet(name string, diags *diagnostic.Diagnostics) *proxmox.Client {
	if client := cm.get(name); client != nil {
		return client
	}

	diags.Append(
		diagnostic.Warning(
			"Client Not Found",
			fmt.Sprintf("no client found with name %q", name),
			diagnostic.WithCategory(diagnostic.Category_InvalidParameter),
		),
	)
	return nil
}

func (cm *ClientManager) Add(name string, client *proxmox.Client, diags *diagnostic.Diagnostics) bool {
	results := defaultClientValidator(client)
	diags.Append(results.Snapshot()...)
	if results.AnySeverity(diagnostic.SEVERITY_Error) {
		return false
	}

	success := cm.add(name, client)
	if !success {
		diags.Append(
			diagnostic.Error(
				"Client Already Exists",
				fmt.Sprintf("a client already exists with the name %q", name),
				diagnostic.WithoutStacktrace(),
				diagnostic.WithCategory(diagnostic.Category_InvalidParameter),
			),
		)
	}

	return success
}

func (cm *ClientManager) Set(name string, client *proxmox.Client, diags *diagnostic.Diagnostics) {
	results := defaultClientValidator(client)
	diags.Append(results.Snapshot()...)
	if results.AnySeverity(diagnostic.SEVERITY_Error) {
		return
	}

	cm.set(name, client)
}
