const KYCModule = {
  checkStatus() {
    if (!App.state.token) {
      document.getElementById('kycGuestNotice').style.display = 'block';
      document.getElementById('kycFormContainer').style.display = 'none';
      return;
    }
    document.getElementById('kycGuestNotice').style.display = 'none';
    document.getElementById('kycFormContainer').style.display = 'block';
  },

  async uploadDocument(e) {
    e.preventDefault();
    const docType = document.getElementById('docTypeSelect').value;
    const fileInput = document.getElementById('docFileInput');

    if (!fileInput.files.length) {
      alert('لطفاً فایل سند را انتخاب کنید.');
      return;
    }

    const formData = new FormData();
    formData.append('document_type', docType);
    formData.append('file', fileInput.files[0]);

    try {
      const res = await App.api('/api/v1/kyc/upload', {
        method: 'POST',
        body: formData
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'خطا در بارگذاری مدرک');

      alert('مدرک با موفقیت بارگذاری و هش امنیتی SHA-256 ثبت گردید.');
      fileInput.value = '';
    } catch (err) {
      alert(err.message);
    }
  }
};
