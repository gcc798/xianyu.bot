import { CompassOutlined, HomeOutlined, ReloadOutlined } from "@ant-design/icons";
import { Button } from "antd";
import { isRouteErrorResponse, useRouteError } from "react-router-dom";
import styles from "./RouteErrorPage.module.css";

const chunkErrorPattern =
  /Failed to fetch dynamically imported module|Importing a module script failed|Loading chunk .* failed/i;

function ErrorContent({ description }: { description: string }) {
  return (
    <main className={styles.page}>
      <header className={styles.header}>闲鱼运营</header>
      <section className={styles.content} aria-labelledby="route-error-title">
        <div className={styles.status} aria-hidden="true">
          <CompassOutlined />
        </div>
        <h1 id="route-error-title">页面走丢了</h1>
        <p>{description}</p>
        <div className={styles.actions}>
          <Button type="primary" icon={<ReloadOutlined />} onClick={() => window.location.reload()}>
            刷新页面
          </Button>
          <Button
            icon={<HomeOutlined />}
            onClick={() => window.location.assign(`${import.meta.env.BASE_URL}overview`)}
          >
            返回总览
          </Button>
        </div>
      </section>
    </main>
  );
}

/** RouteErrorPage 为路由加载错误提供统一恢复入口。 */
export default function RouteErrorPage() {
  const routeError = useRouteError();
  const missingPage = isRouteErrorResponse(routeError) && routeError.status === 404;
  const errorMessage = routeError instanceof Error ? routeError.message : "";
  const versionChanged = chunkErrorPattern.test(errorMessage);
  const description = missingPage
    ? "这个地址可能已经失效。你可以刷新页面，或者返回总览继续操作。"
    : versionChanged
      ? "系统可能刚刚完成更新。刷新页面后会自动加载最新版本。"
      : "页面暂时无法打开。请刷新后重试，仍有问题时返回总览继续操作。";

  return <ErrorContent description={description} />;
}

/** NotFoundPage 为未知地址提供恢复入口。 */
export function NotFoundPage() {
  return (
    <ErrorContent description="这个地址可能已经失效。你可以刷新页面，或者返回总览继续操作。" />
  );
}
