import { Alert, App, Button, Form, Input, InputNumber, Modal, Space, Switch, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { systemApi } from "../../../api/system";
import type { Dictionary, DictionaryDetail, DictionaryDetailInput, DictionaryInput } from "../../../models/system";
import styles from "../../../styles/page.module.css";

export function Component() {
  const { message, modal } = App.useApp();
  const [dictionaryForm] = Form.useForm<DictionaryInput>();
  const [detailForm] = Form.useForm<DictionaryDetailInput>();
  const [dictionaries, setDictionaries] = useState<Dictionary[]>([]);
  const [selectedID, setSelectedID] = useState(0);
  const [details, setDetails] = useState<DictionaryDetail[]>([]);
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [editingDictionary, setEditingDictionary] = useState<Dictionary | null | undefined>(undefined);
  const [editingDetail, setEditingDetail] = useState<DictionaryDetail | null | undefined>(undefined);

  const selected = dictionaries.find((dictionary) => dictionary.ID === selectedID) || null;

  const loadDictionaries = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await systemApi.dictionaries(query);
      setDictionaries(response);
      setSelectedID((current) => response.some((row) => row.ID === current) ? current : response[0]?.ID || 0);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "字典加载失败");
    } finally {
      setLoading(false);
    }
  }, [query]);

  const loadDetails = useCallback(async () => {
    if (!selectedID) {
      setDetails([]);
      return;
    }
    setDetailLoading(true);
    try {
      setDetails((await systemApi.dictionaryDetails(selectedID)).list);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "字典项加载失败");
    } finally {
      setDetailLoading(false);
    }
  }, [selectedID]);

  useEffect(() => { void loadDictionaries(); }, [loadDictionaries]);
  useEffect(() => { void loadDetails(); }, [loadDetails]);

  const openDictionary = (dictionary: Dictionary | null) => {
    setEditingDictionary(dictionary);
    dictionaryForm.setFieldsValue(dictionary ? {
      ID: dictionary.ID,
      name: dictionary.name,
      type: dictionary.type,
      status: dictionary.status,
      desc: dictionary.desc,
      parentID: dictionary.parentID,
    } : { status: true, parentID: null, desc: "" });
  };

  const saveDictionary = async () => {
    const input = await dictionaryForm.validateFields();
    setSaving(true);
    try {
      if (editingDictionary) await systemApi.updateDictionary({ ...input, ID: editingDictionary.ID });
      else await systemApi.createDictionary(input);
      message.success("字典已保存");
      setEditingDictionary(undefined);
      dictionaryForm.resetFields();
      await loadDictionaries();
    } finally {
      setSaving(false);
    }
  };

  const openDetail = (detail: DictionaryDetail | null) => {
    if (!selectedID) return;
    setEditingDetail(detail);
    detailForm.setFieldsValue(detail ? {
      ID: detail.ID,
      label: detail.label,
      value: detail.value,
      extend: detail.extend,
      status: detail.status,
      sort: detail.sort,
      sysDictionaryID: selectedID,
      parentID: detail.parentID,
    } : { status: true, sort: 0, extend: "", sysDictionaryID: selectedID, parentID: null });
  };

  const saveDetail = async () => {
    const input = await detailForm.validateFields();
    setSaving(true);
    try {
      if (editingDetail) await systemApi.updateDictionaryDetail({ ...input, ID: editingDetail.ID });
      else await systemApi.createDictionaryDetail(input);
      message.success("字典项已保存");
      setEditingDetail(undefined);
      detailForm.resetFields();
      await loadDetails();
    } finally {
      setSaving(false);
    }
  };

  const removeDictionary = (dictionary: Dictionary) => modal.confirm({
    title: "删除字典？",
    content: `字典“${dictionary.name}”及其关联结构将无法继续使用。`,
    okText: "删除",
    okType: "danger",
    cancelText: "取消",
    onOk: async () => {
      await systemApi.deleteDictionary(dictionary.ID);
      message.success("字典已删除");
      await loadDictionaries();
    },
  });

  const removeDetail = (detail: DictionaryDetail) => modal.confirm({
    title: "删除字典项？",
    content: `确认删除“${detail.label}”？`,
    okText: "删除",
    okType: "danger",
    cancelText: "取消",
    onOk: async () => {
      await systemApi.deleteDictionaryDetail(detail.ID);
      message.success("字典项已删除");
      await loadDetails();
    },
  });

  const columns: TableColumnsType<DictionaryDetail> = [
    { title: "展示值", dataIndex: "label", width: 180 },
    { title: "字典值", dataIndex: "value", width: 180 },
    { title: "扩展值", dataIndex: "extend", ellipsis: true, render: (value: string) => value || "—" },
    { title: "排序", dataIndex: "sort", width: 80 },
    { title: "状态", dataIndex: "status", width: 90, render: (value: boolean | null) => <Tag color={value ? "success" : "default"}>{value ? "启用" : "停用"}</Tag> },
    { title: "操作", key: "actions", width: 120, render: (_, detail) => <Space size={4}><Button type="link" className={styles.linkButton} onClick={() => openDetail(detail)}>编辑</Button><Button type="link" danger className={styles.linkButton} onClick={() => removeDetail(detail)}>删除</Button></Space> },
  ];

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.filters}>
          <Input.Search value={keyword} allowClear placeholder="字典名称" style={{ width: 280 }} onChange={(event) => setKeyword(event.target.value)} onSearch={() => setQuery(keyword.trim())} />
          <Button onClick={() => { setKeyword(""); setQuery(""); }}>重置</Button>
          <Button type="primary" onClick={() => openDictionary(null)}>新建字典</Button>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.split}>
          <aside className={styles.dictionaryList} aria-busy={loading}>
            {dictionaries.map((dictionary) => (
              <button key={dictionary.ID} data-selected={dictionary.ID === selectedID} onClick={() => setSelectedID(dictionary.ID)}>
                <span>{dictionary.name}</span>
                <small>{dictionary.type}</small>
              </button>
            ))}
          </aside>
          <section className={styles.dictionaryDetails}>
            <div className={styles.sectionToolbar}>
              <strong>{selected?.name || "字典项"}</strong>
              {selected && <Space><Button onClick={() => openDictionary(selected)}>编辑字典</Button><Button danger onClick={() => removeDictionary(selected)}>删除字典</Button><Button type="primary" onClick={() => openDetail(null)}>新增字典项</Button></Space>}
            </div>
            <div className={styles.table}>
              <Table<DictionaryDetail> rowKey="ID" size="small" loading={detailLoading} columns={columns} dataSource={details} pagination={false} scroll={{ x: 760 }} />
            </div>
          </section>
        </div>
      </section>

      <Modal title={editingDictionary ? "编辑字典" : "新建字典"} open={editingDictionary !== undefined} confirmLoading={saving} okText="保存" cancelText="取消" destroyOnHidden onOk={() => void saveDictionary()} onCancel={() => { setEditingDictionary(undefined); dictionaryForm.resetFields(); }}>
        <Form<DictionaryInput> form={dictionaryForm} layout="vertical" requiredMark={false}>
          <Form.Item name="name" label="字典名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="type" label="字典类型" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="desc" label="说明"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="parentID" label="父字典 ID"><InputNumber min={1} style={{ width: "100%" }} /></Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked"><Switch /></Form.Item>
        </Form>
      </Modal>

      <Modal title={editingDetail ? "编辑字典项" : "新增字典项"} open={editingDetail !== undefined} confirmLoading={saving} okText="保存" cancelText="取消" destroyOnHidden onOk={() => void saveDetail()} onCancel={() => { setEditingDetail(undefined); detailForm.resetFields(); }}>
        <Form<DictionaryDetailInput> form={detailForm} layout="vertical" requiredMark={false}>
          <Form.Item name="label" label="展示值" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="value" label="字典值" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="extend" label="扩展值"><Input /></Form.Item>
          <Space size={12} style={{ display: "flex" }}>
            <Form.Item name="sort" label="排序" style={{ flex: 1 }}><InputNumber style={{ width: "100%" }} /></Form.Item>
            <Form.Item name="parentID" label="父项 ID" style={{ flex: 1 }}><InputNumber min={1} style={{ width: "100%" }} /></Form.Item>
          </Space>
          <Form.Item name="status" label="启用" valuePropName="checked"><Switch /></Form.Item>
          <Form.Item name="sysDictionaryID" hidden><InputNumber /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
