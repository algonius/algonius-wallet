import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./index.css"; // 预留 Tailwind 导入

const container = document.getElementById("root");
if (container) {
  const root = createRoot(container);
  root.render(<App />);
}
