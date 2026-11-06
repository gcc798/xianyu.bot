import { Alert, App, Button, Form, Input, Modal, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { systemApi } from "../../../api/system";
import { ADMIN_AUTHORITY_ID, USER_AUTHORITY_ID } from "../../../models/auth";
import type { CreateUserRequest, UpdateUserRequest, UserRecord } from "../../../models/system";
import styles from "../../../styles/page.module.css";

type UserForm = {
  userName: string;
  passWord: string;
  nickName: string;
  phone: string;
  email: string;
  headerImg: string;
  authorityId: number;
  enable: number;
};

export function Component() {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<UserForm>();
  const [rows, setRows] = useState<UserRecord[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<UserRecord | null | undefined>(undefined);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await systemApi.users({
        page,
        pageSize: 20,
        username: query,
        nickName: query,
        orderKey: "id",
        desc: true,
      });
      setRows(response.list);
      setTotal(response.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "用户列表加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, query]);

  useEffect(() => { void load(); }, [load]);

  const openCreate = () => {
    setEditing(null);
    form.setFieldsValue({ authorityId: USER_AUTHORITY_ID, enable: 1, headerImg: "" });
  };

  const openEdit = (user: UserRecord) => {
    setEditing(user);
    form.setFieldsValue({
      userName: user.userName,
      passWord: "",
      nickName: user.nickName,
      phone: user.phone,
      email: user.email,
      headerImg: user.headerImg,
      authorityId: user.authorityId,
      enable: user.enable,
    });
  };

  const save = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      if (editing) {
        const input: UpdateUserRequest = {
          ID: editing.ID,
          nickName: values.nickName,
          phone: values.phone || "",
          email: values.email || "",
          headerImg: values.headerImg || "",
          authorityId: values.authorityId,
          enable: values.enable,
        };
        await systemApi.updateUser(input);
      } else {
        const input: CreateUserRequest = {
          userName: values.userName,
          passWord: values.passWord,
          nickName: values.nickName,
          phone: values.phone || "",
          email: values.email || "",
          headerImg: values.headerImg || "",
          authorityId: values.authorityId,
          enable: values.enable,
        };
        await systemApi.createUser(input);
      }
      message.success("用户已保存");
      setEditing(undefined);
      form.resetFields();
      await load();
    } finally {
      setSaving(false);
    }
  };

  const remove = (user: UserRecord) => {
    modal.confirm({
      title: "删除用户？",
      content: `删除后 ${user.userName} 将无法登录。`,
      okText: "删除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        await systemApi.deleteUser(user.ID);
        message.success("用户已删除");
        await load();
      },
    });
  };

  const resetPassword = (user: UserRecord) => {
    let password = "";
    modal.confirm({
      title: `重置 ${user.userName} 的密码`,
      content: <Input.Password autoComplete="new-password" placeholder="至少 6 位" onChange={(event) => { password = event.target.value; }} />,
      okText: "重置",
      cancelText: "取消",
      onOk: async () => {
        if (password.length < 6) throw new Error("密码至少 6 位");
        await systemApi.resetPassword(user.ID, password);
        message.success("密码已重置");
      },
    });
  };

  const columns: TableColumnsType<UserRecord> = [
    { title: "序号", width: 64, align: "center", render: (_value, _record, index) => (page - 1) * 20 + index + 1 },
    { title: "用户名", dataIndex: "userName", width: 160 },
    { title: "昵称", dataIndex: "nickName", width: 160 },
    { title: "角色", dataIndex: "authorityId", width: 110, render: (value: number) => value === ADMIN_AUTHORITY_ID ? "管理员" : "普通用户" },
    { title: "手机号", dataIndex: "phone", width: 150, render: (value: string) => value || "—" },
    { title: "邮箱", dataIndex: "email", ellipsis: true, render: (value: string) => value || "—" },
    { title: "状态", dataIndex: "enable", width: 96, render: (value: number) => <Tag color={value === 1 ? "success" : "default"}>{value === 1 ? "启用" : "停用"}</Tag> },
    {
      title: "操作",
      key: "actions",
      width: 200,
      render: (_, user) => (
        <Space size={4}>
          <Button type="link" className={styles.linkButton} onClick={() => openEdit(user)}>编辑</Button>
          <Button type="link" className={styles.linkButton} onClick={() => resetPassword(user)}>重置密码</Button>
          <Button type="link" danger className={styles.linkButton} onClick={() => remove(user)}>删除</Button>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.filters}>
          <Input.Search
            value={keyword}
            allowClear
            placeholder="用户名或昵称"
            style={{ width: 280 }}
            onChange={(event) => setKeyword(event.target.value)}
            onSearch={() => { setPage(1); setQuery(keyword.trim()); }}
          />
          <Button onClick={() => { setKeyword(""); setQuery(""); setPage(1); }}>重置</Button>
          <Button type="primary" onClick={openCreate}>新建用户</Button>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<UserRecord> rowKey="ID" size="small" loading={loading} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 980 }} />
        </div>
        <div className={styles.pagination}>
          <Pagination current={page} pageSize={20} total={total} showSizeChanger={false} onChange={setPage} />
        </div>
      </section>

      <Modal
        title={editing ? "编辑用户" : "新建用户"}
        open={editing !== undefined}
        confirmLoading={saving}
        okText="保存"
        cancelText="取消"
        destroyOnHidden
        onOk={() => void save()}
        onCancel={() => { setEditing(undefined); form.resetFields(); }}
      >
        <Form<UserForm> form={form} layout="vertical" requiredMark={false}>
          <Form.Item name="userName" label="用户名" rules={[{ required: true }]}><Input disabled={Boolean(editing)} autoComplete="off" /></Form.Item>
          {!editing && <Form.Item name="passWord" label="初始密码" rules={[{ required: true }, { min: 6 }]}><Input.Password autoComplete="new-password" /></Form.Item>}
          <Form.Item name="nickName" label="昵称" rules={[{ required: true }]}><Input /></Form.Item>
          <Space size={12} style={{ display: "flex" }}>
            <Form.Item name="authorityId" label="角色" style={{ flex: 1 }}><Select options={[{ label: "普通用户", value: USER_AUTHORITY_ID }, { label: "管理员", value: ADMIN_AUTHORITY_ID }]} /></Form.Item>
            <Form.Item name="enable" label="状态" style={{ flex: 1 }}><Select options={[{ label: "启用", value: 1 }, { label: "停用", value: 2 }]} /></Form.Item>
          </Space>
          <Form.Item name="phone" label="手机号"><Input /></Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: "email" }]}><Input /></Form.Item>
          <Form.Item name="headerImg" label="头像地址"><Input /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
