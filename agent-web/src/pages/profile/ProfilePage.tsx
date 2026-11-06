import { App, Button, Form, Input, Tabs } from "antd";
import { useEffect, useState } from "react";
import { changePassword, updateCurrentUser } from "../../api/auth";
import { useAuth } from "../../app/AuthProvider";
import pageStyles from "../../styles/page.module.css";
import styles from "./ProfilePage.module.css";

type ProfileForm = { userName: string; nickName: string; phone: string; email: string };
type PasswordForm = { password: string; newPassword: string; confirmPassword: string };

export function Component() {
  const { message } = App.useApp();
  const { session, refreshUser } = useAuth();
  const [profileForm] = Form.useForm<ProfileForm>();
  const [passwordForm] = Form.useForm<PasswordForm>();
  const [savingProfile, setSavingProfile] = useState(false), [savingPassword, setSavingPassword] = useState(false);

  useEffect(() => {
    if (session) profileForm.setFieldsValue({
      userName: session.user.userName,
      nickName: session.user.nickName,
      phone: session.user.phone,
      email: session.user.email,
    });
  }, [profileForm, session]);

  const saveProfile = async (values: ProfileForm) => {
    if (!session) return;
    setSavingProfile(true);
    try {
      await updateCurrentUser({ nickName: values.nickName.trim(), phone: values.phone.trim(), email: values.email.trim(), headerImg: session.user.headerImg });
      await refreshUser();
      message.success("个人资料已保存");
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "个人资料保存失败");
    } finally {
      setSavingProfile(false);
    }
  };

  const savePassword = async (values: PasswordForm) => {
    setSavingPassword(true);
    try {
      await changePassword({ password: values.password, newPassword: values.newPassword });
      passwordForm.resetFields();
      message.success("密码已修改");
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "密码修改失败");
    } finally {
      setSavingPassword(false);
    }
  };

  const items = [
    { key: "profile", label: "个人资料", children: <Form<ProfileForm> form={profileForm} layout="vertical" className={styles.form} requiredMark={false} onFinish={saveProfile}>
      <div className={styles.row}>
        <Form.Item name="userName" label="用户名"><Input disabled /></Form.Item>
        <Form.Item name="nickName" label="昵称" rules={[{ required: true, whitespace: true, message: "请输入昵称" }, { max: 64 }]}><Input maxLength={64} /></Form.Item>
      </div>
      <div className={styles.row}>
        <Form.Item name="phone" label="手机号"><Input maxLength={32} /></Form.Item>
        <Form.Item name="email" label="邮箱" rules={[{ type: "email", message: "请输入有效邮箱" }]}><Input maxLength={128} /></Form.Item>
      </div>
      <Button className={styles.actions} type="primary" htmlType="submit" loading={savingProfile}>保存资料</Button>
    </Form> },
    { key: "password", label: "修改密码", children: <Form<PasswordForm> form={passwordForm} layout="vertical" className={styles.form} requiredMark={false} onFinish={savePassword}>
      <Form.Item name="password" label="当前密码" rules={[{ required: true }, { min: 6 }]}><Input.Password autoComplete="current-password" /></Form.Item>
      <Form.Item name="newPassword" label="新密码" rules={[{ required: true }, { min: 6, message: "密码至少 6 位" }]}><Input.Password autoComplete="new-password" /></Form.Item>
      <Form.Item name="confirmPassword" label="确认新密码" dependencies={["newPassword"]} rules={[{ required: true }, ({ getFieldValue }) => ({ validator: (_, value) => !value || value === getFieldValue("newPassword") ? Promise.resolve() : Promise.reject(new Error("两次输入的密码不一致")) })]}><Input.Password autoComplete="new-password" /></Form.Item>
      <Button className={styles.actions} type="primary" htmlType="submit" loading={savingPassword}>修改密码</Button>
    </Form> },
  ];

  return <div className={pageStyles.page}><section className={pageStyles.surface}><div className={styles.content}><Tabs items={items} /></div></section></div>;
}
