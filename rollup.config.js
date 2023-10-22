import { nodeResolve } from "@rollup/plugin-node-resolve";
import terser from "@rollup/plugin-terser";
import typescript from "@rollup/plugin-typescript";

export default [
  {
    input: "script/index.ts",
    output: [
      {
        dir: "dist/scripts",
        format: "iife",
        // plugins: [terser()]
      },
    ],
    plugins: [typescript()],
  },
  {
    input: ["script/modules/auth.ts"],
    output: [
      {
        dir: "dist/scripts",
        format: "iife",
        // plugins: [terser()]
      },
    ],
    plugins: [nodeResolve(), typescript()],
  },
];
