import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import ProductManagementPreview from "./previews/product-management";

// 预览入口与正式 main.tsx 完全隔离，不读取登录态，也不调用业务 API。
createRoot(document.getElementById("preview-root")!).render(
  <StrictMode>
    <ProductManagementPreview />
  </StrictMode>
);
