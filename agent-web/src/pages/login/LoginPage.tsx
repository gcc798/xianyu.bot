import { Alert, Button, Form, Input } from "antd";
import { useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../app/AuthContext";
import type { LoginRequest } from "../../models/auth";
import styles from "./LoginPage.module.css";

export function Component() {
  const { session, signIn } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  if (session) return <Navigate to="/" replace />;

  const submit = async (input: LoginRequest) => {
    setSubmitting(true);
    setError("");
    try {
      await signIn(input);
      navigate((location.state as { from?: string } | null)?.from || "/", { replace: true });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "登录失败，请重试");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <main className={styles.page}>
      <section className={styles.panel}>
        <header>
          <h1>闲鱼运营</h1>
        </header>
        {error && <Alert type="error" showIcon title={error} />}
        <Form<LoginRequest>
          layout="vertical"
          requiredMark={false}
          onFinish={(values) => void submit(values)}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: "请输入用户名" }]}
          >
            <Input autoFocus autoComplete="username" />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[{ required: true, message: "请输入密码" }]}
          >
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={submitting}>
            登录
          </Button>
        </Form>
      </section>
    </main>
  );
}
