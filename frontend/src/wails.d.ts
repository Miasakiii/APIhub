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
        }
      }
    }
  }
}
