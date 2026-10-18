package app

import "github.com/Uzer007-admin/granio/v2"

func (a *fyneApp) SetCloudProvider(p granio.CloudProvider) {
	if p == nil {
		a.cloud = nil
		return
	}

	a.transitionCloud(p)
}

func (a *fyneApp) transitionCloud(p granio.CloudProvider) {
	if a.cloud != nil {
		a.cloud.Cleanup(a)
	}

	err := p.Setup(a)
	if err != nil {
		granio.LogError("Failed to set up cloud provider "+p.ProviderName(), err)
		return
	}
	a.cloud = p

	listeners := a.prefs.ChangeListeners()
	if pp, ok := p.(granio.CloudProviderPreferences); ok {
		a.prefs = pp.CloudPreferences(a)
	} else {
		a.prefs = a.newDefaultPreferences()
	}
	if cloud, ok := p.(granio.CloudProviderStorage); ok {
		a.storage = cloud.CloudStorage(a)
	} else {
		store := &store{a: a}
		store.Docs = makeStoreDocs(a.uniqueID, store)
		a.storage = store
	}

	for _, l := range listeners {
		a.prefs.AddChangeListener(l)
		l() // assume that preferences have changed because we replaced the provider
	}

	// after transition ensure settings listener is fired
	a.settings.apply()
}
