export {}

declare global {
  interface Window {
    go?: {
      main?: {
        WailsApp?: {
          GetAPIPort(): Promise<string>
          GetAPIURL(): Promise<string>
          GetVersion(): Promise<string>
          GetDataDir(): Promise<string>
          OpenExternalURL(url: string): Promise<void>
          MinimizeToTray(): Promise<void>
          ShowWindow(): Promise<void>
          SetMinimizeToTray(enable: boolean): Promise<void>
          GetMinimizeToTray(): Promise<boolean>
          ShowNotification(title: string, message: string): Promise<void>
          SetAutoStart(enable: boolean): Promise<void>
          IsAutoStartEnabled(): Promise<boolean>
        }
      }
    }
  }
}
