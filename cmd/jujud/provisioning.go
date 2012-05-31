package main

import (
	"launchpad.net/gnuflag"
	"launchpad.net/juju/go/cmd"
	"launchpad.net/juju/go/environs"
	"launchpad.net/juju/go/log"
	"launchpad.net/juju/go/state"
	"launchpad.net/tomb"

	// register providers
	_ "launchpad.net/juju/go/environs/dummy"
	_ "launchpad.net/juju/go/environs/ec2"
)

// ProvisioningAgent is a cmd.Command responsible for running a provisioning agent.
type ProvisioningAgent struct {
	Conf AgentConf
}

// Info returns usage information for the command.
func (a *ProvisioningAgent) Info() *cmd.Info {
	return &cmd.Info{"provisioning", "", "run a juju provisioning agent", ""}
}

// Init initializes the command for running.
func (a *ProvisioningAgent) Init(f *gnuflag.FlagSet, args []string) error {
	a.Conf.addFlags(f)
	if err := f.Parse(true, args); err != nil {
		return err
	}
	return a.Conf.checkArgs(f.Args())
}

// Run runs a provisioning agent.
func (a *ProvisioningAgent) Run(_ *cmd.Context) error {
	st, err := state.Open(&a.Conf.StateInfo)
	if err != nil {
		return err
	}
	p := NewProvisioner(st)
	return p.tomb.Wait()
}

type Provisioner struct {
	st      *state.State
	environ environs.Environ
	tomb    tomb.Tomb

	environment
	machines
}

// environment ensures that the watcher for the environ
// configuration is valid.
type environment struct {
	st      *state.State
	watcher *state.ConfigWatcher
}

// changes returns a channel that will receive the new *ConfigNode when a
// change is detected. 
func (e *environment) changes() <-chan *state.ConfigNode {
	if e.watcher == nil {
		e.watcher = e.st.WatchEnvironConfig()
	}
	return e.watcher.Changes()
}

// invalidate stops the current watcher.
func (e *environment) invalidate() {
	if e.watcher != nil {
		log.Printf("provisioner: environment watcher exited: %v", e.watcher.Stop())
	}
	e.watcher = nil
}

// machines ensures that the watcher for machines changes
// is valid.
type machines struct {
	st      *state.State
	watcher *state.MachinesWatcher
}

// changes returns a channel that will receive the new *ConfigNode when a
// change is detected. 
func (m *machines) changes() <-chan *state.MachinesChange {
	if m.watcher == nil {
		m.watcher = m.st.WatchMachines()
	}
	return m.watcher.Changes()
}

// invalidate stops the current watcher.
func (m *machines) invalidate() {
	if m.watcher != nil {
		log.Printf("provisioner: machines watcher exited: %v", m.watcher.Stop())
	}
	m.watcher = nil
}

// NewProvisioner returns a Provisioner.
func NewProvisioner(st *state.State) *Provisioner {
	p := &Provisioner{
		st:          st,
		environment: environment{st: st},
		machines:    machines{st: st},
	}
	go p.loop()
	return p
}

func (p *Provisioner) loop() {
	defer p.tomb.Done()
	for {
		select {
		case <-p.tomb.Dying():
			return
		case config, ok := <-p.environment.changes():
			if !ok {
				p.environment.invalidate()
				continue
			}
			var err error
			p.environ, err = environs.NewEnviron(config.Map())
			if err != nil {
				log.Printf("provisioner: unable to create environment from supplied configuration: %v", err)
				continue
			}
			log.Printf("provisioning: valid environment configured")
			p.innerLoop()
		}
	}
}

func (p *Provisioner) innerLoop() {
	for {
		select {
		case <-p.tomb.Dying():
			return
		case change, ok := <-p.environment.changes():
			if !ok {
				p.environment.invalidate()
				continue
			}
			config, err := environs.NewConfig(change.Map())
			if err != nil {
				log.Printf("provisioning: new configuration received, but was not valid: %v", err)
				continue
			}
			p.environ.SetConfig(config)
			log.Printf("provisioning: new configuartion applied")
		case machines, ok := <-p.machines.changes():
			if !ok {
				p.machines.invalidate()
				continue
			}
			p.processMachines(machines)
		}
	}
}

// Stop stops the Provisioner and returns any error encountered while
// provisioning.
func (p *Provisioner) Stop() error {
	p.tomb.Kill(nil)
	p.environment.invalidate()
	p.machines.invalidate()
	return p.tomb.Wait()
}

func (p *Provisioner) processMachines(changes *state.MachinesChange) {}
