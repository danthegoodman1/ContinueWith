"use client"

import Image from 'next/image'
import { useEffect, useState } from 'react'

function getCookie(cname: string) {
  let name = cname + "=";
  let decodedCookie = decodeURIComponent(document.cookie);
  let ca = decodedCookie.split(';');
  for(let i = 0; i <ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) == ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}

export default function Home() {
  const [userName, setUserName] = useState<string | undefined>()
  const [response, setResponse] = useState<{code: number; text: string} | undefined>()

  useEffect(() => {
    const user = getCookie("user");
    (async () => {
      const res = await fetch("/api/user")
      setResponse({
        code: res.status,
        text: await res.text()
      })
    })();
    if (!user) {
      return
    }

    setUserName(user);
  }, [])

  return (
    <main>
      <h1>Logged in: {userName || "(not logged in)"}</h1>
      <p>API response: {response?.code} {response?.text}</p>
    </main>
  )
}
