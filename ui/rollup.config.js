import typescript from "@rollup/plugin-typescript"
import nodeResolve from "@rollup/plugin-node-resolve"
import scss from "rollup-plugin-scss"

export default {
  input: "@vpui/vephar.tsx",
  output: {dir: "../srv/dist", format: "esm", sourcemap: "inline"},
  plugins: [
    nodeResolve(),
    typescript(),
    scss({output: "../srv/dist/vephar.css", failOnError: true})
  ]
}
