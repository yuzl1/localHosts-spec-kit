export function getHosts() {
  return window.go.main.App.GetHosts()
}

export function canWriteHosts() {
  return window.go.main.App.CanWriteHosts()
}

export function ensureWriteAccess() {
  return window.go.main.App.EnsureWriteAccess()
}

export function saveHosts(workspace) {
  return window.go.main.App.SaveHosts(workspace)
}

export function applyHosts(workspace) {
  return window.go.main.App.ApplyHosts(workspace)
}
