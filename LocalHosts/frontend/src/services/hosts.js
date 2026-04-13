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

export function getAppVersion() {
  return window.go.main.App.GetAppVersion()
}

export function getUpdateSettings() {
  return window.go.main.App.GetUpdateSettings()
}

export function saveUpdateSettings(settings) {
  return window.go.main.App.SaveUpdateSettings(settings)
}

export function checkForUpdate() {
  return window.go.main.App.CheckForUpdate()
}

export function downloadUpdate(info) {
  return window.go.main.App.DownloadUpdate(info)
}

export function installUpdate(installerPath) {
  return window.go.main.App.InstallUpdate(installerPath)
}
