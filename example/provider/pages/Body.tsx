import React from "preact/compat";

export default function Body(props: React.PropsWithChildren) {
  return <html>
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
    <body>{props.children}</body>
  </html>
}
