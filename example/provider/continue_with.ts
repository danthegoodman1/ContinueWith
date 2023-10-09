export interface Client {
  ID: string
  Name: string
  Suspended: boolean
  Created: string
  Updated: string
}

export async function GetClient(clientID: string): Promise<Client> {
  const res = await fetch(
    process.env.CW_API + "/admin/client/" + encodeURIComponent(clientID),
    {
      headers: {
        Authorization: `Bearer ${process.env.CW_ADMIN_KEY}`,
      },
    }
  )
  if (res.status > 299) {
    throw new Error(`Status ${res.status}: ${await res.text()}`)
  }

  return await res.json()
}

export interface AccessToken {
  UserID: string
  CreatedMS: number
  ExpiresMS: number
  Scopes: string[]
}

export async function GetAccessToken(token: string): Promise<AccessToken> {
  const res = await fetch(
    process.env.CW_API + "/admin/access_token/" + encodeURIComponent(token),
    {
      headers: {
        Authorization: `Bearer ${process.env.CW_ADMIN_KEY}`,
      },
    }
  )
  if (res.status > 299) {
    throw new Error(`Status ${res.status}: ${await res.text()}`)
  }

  return await res.json()
}
