# MinIO 上传预签名接口文档

## 概述

本文档说明如何使用 MinIO 预签名接口上传简历文件。提供了两种上传方式：
1. **PUT 预签名 URL**：直接使用 HTTP PUT 上传
2. **POST Policy**：使用表单 POST 上传

---

## 配置

### MinIO 配置 (etc/cv2.yaml)

```yaml
MinIO:
  Endpoint: localhost:9000
  AccessKeyID: minioadmin
  SecretAccessKey: minioadmin
  UseSSL: false
  BucketName: cv2-resumes
```

---

## API 接口

### 1. 获取 PUT 预签名 URL

**接口地址**：`POST /api/resume/upload/presign`

**请求头**：
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**：
```json
{
  "file_name": "resume.pdf",
  "file_type": "pdf"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file_name | string | 是 | 文件名 |
| file_type | string | 是 | 文件类型（pdf, docx, doc 等） |

**响应示例**：
```json
{
  "upload_url": "http://localhost:9000/cv2-resumes/resumes/1234567890.pdf?X-Amz-Algorithm=...",
  "object_key": "resumes/1234567890.pdf",
  "expires_in": 900
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| upload_url | string | 预签名上传 URL |
| object_key | string | 对象键（文件在 MinIO 中的路径） |
| expires_in | int64 | 过期时间（秒），默认 900 秒（15分钟） |

**使用示例（JavaScript）**：
```javascript
// 1. 获取预签名 URL
const response = await fetch('/api/resume/upload/presign', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    file_name: 'resume.pdf',
    file_type: 'pdf'
  })
});

const data = await response.json();

// 2. 使用预签名 URL 上传文件
const file = document.getElementById('fileInput').files[0];
await fetch(data.upload_url, {
  method: 'PUT',
  body: file,
  headers: {
    'Content-Type': 'application/pdf'
  }
});

console.log('文件上传成功，对象键:', data.object_key);
```

**使用示例（curl）**：
```bash
# 1. 获取预签名 URL
RESPONSE=$(curl -X POST http://localhost:8888/api/resume/upload/presign \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"file_name":"resume.pdf","file_type":"pdf"}')

UPLOAD_URL=$(echo $RESPONSE | jq -r '.upload_url')

# 2. 上传文件
curl -X PUT "$UPLOAD_URL" \
  -H "Content-Type: application/pdf" \
  --data-binary @resume.pdf
```

---

### 2. 获取 POST Policy（表单上传）

**接口地址**：`POST /api/resume/upload/policy`

**请求头**：
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**：
```json
{
  "file_name": "resume.pdf",
  "file_type": "pdf",
  "max_file_size": 104857600
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file_name | string | 是 | 文件名 |
| file_type | string | 是 | 文件类型 |
| max_file_size | int64 | 否 | 最大文件大小（字节），默认 100MB |

**响应示例**：
```json
{
  "url": "http://localhost:9000/cv2-resumes",
  "form_data": {
    "key": "resumes/1234567890.pdf",
    "policy": "eyJleHBpcmF0aW9uIjoiMjAyNC0w...",
    "x-amz-algorithm": "AWS4-HMAC-SHA256",
    "x-amz-credential": "minioadmin/20240101/us-east-1/s3/aws4_request",
    "x-amz-date": "20240101T000000Z",
    "x-amz-signature": "abcdef123456..."
  },
  "object_key": "resumes/1234567890.pdf",
  "expires_in": 900
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 表单提交 URL |
| form_data | object | 表单字段数据 |
| object_key | string | 对象键 |
| expires_in | int64 | 过期时间（秒） |

**使用示例（JavaScript - FormData）**：
```javascript
// 1. 获取 POST Policy
const response = await fetch('/api/resume/upload/policy', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    file_name: 'resume.pdf',
    file_type: 'pdf',
    max_file_size: 10 * 1024 * 1024 // 10MB
  })
});

const data = await response.json();

// 2. 构建表单数据
const formData = new FormData();

// 添加 policy 字段（必须在文件之前）
Object.keys(data.form_data).forEach(key => {
  formData.append(key, data.form_data[key]);
});

// 添加文件（必须最后添加）
const file = document.getElementById('fileInput').files[0];
formData.append('file', file);

// 3. 上传文件
await fetch(data.url, {
  method: 'POST',
  body: formData
});

console.log('文件上传成功，对象键:', data.object_key);
```

**使用示例（HTML 表单）**：
```html
<!-- 1. 先调用 API 获取 form_data -->
<form id="uploadForm" method="POST" enctype="multipart/form-data">
  <!-- 动态插入 form_data 字段 -->
  <input type="hidden" name="key" value="resumes/1234567890.pdf">
  <input type="hidden" name="policy" value="eyJleHBpcmF0aW9uIjoiMjAyNC0w...">
  <input type="hidden" name="x-amz-algorithm" value="AWS4-HMAC-SHA256">
  <input type="hidden" name="x-amz-credential" value="...">
  <input type="hidden" name="x-amz-date" value="...">
  <input type="hidden" name="x-amz-signature" value="...">
  
  <!-- 文件输入（必须最后） -->
  <input type="file" name="file" required>
  <button type="submit">上传</button>
</form>

<script>
// 动态设置表单 action 和字段
async function setupForm() {
  const response = await fetch('/api/resume/upload/policy', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer ' + token,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      file_name: 'resume.pdf',
      file_type: 'pdf'
    })
  });
  
  const data = await response.json();
  
  // 设置表单 action
  document.getElementById('uploadForm').action = data.url;
  
  // 动态添加隐藏字段
  Object.keys(data.form_data).forEach(key => {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = key;
    input.value = data.form_data[key];
    document.getElementById('uploadForm').insertBefore(
      input,
      document.querySelector('input[type="file"]')
    );
  });
}

setupForm();
</script>
```

---

## 两种方式对比

| 特性 | PUT 预签名 URL | POST Policy |
|------|---------------|-------------|
| 实现复杂度 | 简单 | 中等 |
| 浏览器兼容性 | 好 | 好 |
| 文件大小限制 | 由 MinIO 决定 | 可自定义 |
| 适用场景 | 单文件上传 | 表单上传、多字段 |
| CORS 要求 | 需要配置 | 需要配置 |

---

## 完整上传流程示例

### React 示例

```jsx
import React, { useState } from 'react';
import axios from 'axios';

function ResumeUpload() {
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [objectKey, setObjectKey] = useState('');

  const handleFileChange = (e) => {
    setFile(e.target.files[0]);
  };

  const uploadFile = async () => {
    if (!file) return;

    setUploading(true);
    try {
      // 1. 获取预签名 URL
      const { data } = await axios.post(
        '/api/resume/upload/presign',
        {
          file_name: file.name,
          file_type: file.name.split('.').pop()
        },
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        }
      );

      // 2. 上传文件到 MinIO
      await axios.put(data.upload_url, file, {
        headers: {
          'Content-Type': file.type
        }
      });

      setObjectKey(data.object_key);
      alert('上传成功！');

      // 3. 保存简历记录到数据库
      await axios.post('/api/resume', {
        file_name: file.name,
        file_path: data.object_key
      }, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

    } catch (error) {
      console.error('上传失败:', error);
      alert('上传失败');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input type="file" onChange={handleFileChange} accept=".pdf,.doc,.docx" />
      <button onClick={uploadFile} disabled={!file || uploading}>
        {uploading ? '上传中...' : '上传简历'}
      </button>
      {objectKey && <p>文件路径: {objectKey}</p>}
    </div>
  );
}

export default ResumeUpload;
```

---

## 注意事项

1. **过期时间**：预签名 URL 默认 15 分钟有效，过期后需要重新获取
2. **文件大小**：POST Policy 默认限制 100MB，可通过 `max_file_size` 参数调整
3. **文件类型**：建议限制为 `.pdf`, `.doc`, `.docx` 等简历格式
4. **对象键格式**：自动生成格式为 `resumes/{雪花ID}.{扩展名}`
5. **CORS 配置**：如果前端域名与 MinIO 不同，需要配置 MinIO 的 CORS 策略
6. **安全性**：所有接口都需要 Auth 中间件验证，确保用户已登录

---

## MinIO CORS 配置

如果前端和 MinIO 不在同一域名，需要配置 CORS：

```bash
# 使用 mc 命令行工具配置
mc alias set myminio http://localhost:9000 minioadmin minioadmin

# 设置 bucket 的 CORS 规则
mc anonymous set-json cors.json myminio/cv2-resumes
```

**cors.json**：
```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["http://localhost:3000"],
      "AllowedMethods": ["GET", "PUT", "POST"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"]
    }
  ]
}
```

---

## 错误处理

常见错误及处理：

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| 401 Unauthorized | Token 无效或过期 | 重新登录获取 token |
| 403 Forbidden | MinIO 权限不足 | 检查 AccessKey 和 SecretKey |
| 404 Not Found | Bucket 不存在 | 检查配置或手动创建 bucket |
| 413 Payload Too Large | 文件超过大小限制 | 减小文件或调整 max_file_size |
| 签名过期 | 超过 15 分钟 | 重新获取预签名 URL |

---

## 测试

使用 curl 测试接口：

```bash
# 测试 PUT 预签名
curl -X POST http://localhost:8888/api/resume/upload/presign \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"file_name":"test.pdf","file_type":"pdf"}'

# 测试 POST Policy
curl -X POST http://localhost:8888/api/resume/upload/policy \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"file_name":"test.pdf","file_type":"pdf","max_file_size":10485760}'
```
