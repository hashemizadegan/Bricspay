const AdminModule = {
  async loadData() {
    if (!App.state.token || App.state.user?.role !== 'admin') return;
    this.loadProfiles();
    this.loadAuditLogs();
  },

  async loadProfiles() {
    try {
      const res = await App.api('/api/v1/admin/kyc/list');
      const data = await res.json();
      const tbody = document.getElementById('adminProfilesTableBody');
      tbody.innerHTML = '';

      (data.profiles || []).forEach(p => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td>${p.LegalName}</td>
          <td>${p.Email}</td>
          <td>${p.Jurisdiction}</td>
          <td>${p.Docs} مدرک</td>
          <td><span class="badge badge-${p.Status.toLowerCase()}">${p.Status}</span></td>
          <td>
            <button class="btn btn-success" style="padding:0.2rem 0.5rem;font-size:0.75rem;" onclick="AdminModule.decide('${p.ID}', 'APPROVE')">تایید</button>
            <button class="btn btn-danger" style="padding:0.2rem 0.5rem;font-size:0.75rem;" onclick="AdminModule.decide('${p.ID}', 'REJECT')">رد</button>
          </td>
        `;
        tbody.appendChild(tr);
      });
    } catch (err) {
      console.error('Error fetching admin profiles:', err);
    }
  },

  async decide(profileId, decision) {
    const notes = prompt(`یادداشت بازبینی برای وضعیت ${decision}:`, '');
    if (notes === null) return;

    try {
      const res = await App.api('/api/v1/admin/kyc/decision', {
        method: 'POST',
        body: JSON.stringify({ ProfileID: profileId, Decision: decision, Notes: notes })
      });
      if (!res.ok) throw new Error('خطا در ثبت تصمیم');
      this.loadProfiles();
    } catch (err) {
      alert(err.message);
    }
  },

  async loadAuditLogs() {
    try {
      const res = await App.api('/api/v1/admin/audit');
      const data = await res.json();
      const tbody = document.getElementById('adminAuditTableBody');
      tbody.innerHTML = '';

      (data.audit || []).forEach(a => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td><code>${a.Action}</code></td>
          <td>${a.Actor}</td>
          <td>${a.Entity} (${a.TargetID || '-'})</td>
          <td>${a.IP}</td>
          <td style="font-size:0.75rem;">${new Date(a.Time).toLocaleString('fa-IR')}</td>
        `;
        tbody.appendChild(tr);
      });
    } catch (err) {
      console.error('Error fetching audit logs:', err);
    }
  }
};
