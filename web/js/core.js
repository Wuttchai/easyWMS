const $=id=>document.getElementById(id);let token=localStorage.getItem('easywms_token')||'';let me=JSON.parse(localStorage.getItem('easywms_user')||'null');let activeMaster='products',scanTarget='',scanStream=null,scanTimer=null;
const titles={dashboard:['Dashboard','ภาพรวมคลังสินค้าแบบ Real-time'],master:['Master Data','ข้อมูลหลักของคลังสินค้า'],inventory:['Inventory','Stock แยกตาม Location'],lots:['Lot / Batch / Expiry','ติดตาม Stock ตาม Lot และวันหมดอายุ'],receive:['Receive','ประวัติการรับสินค้า'],issue:['Issue','ประวัติการเบิกสินค้า'],transfer:['Stock Transfer','ประวัติการย้ายสินค้า'],stockcount:['Stock Count','ประวัติการตรวจนับสินค้า'],adjustment:['Adjustment Approval','อนุมัติ/ปฏิเสธการปรับ Stock'],movements:['Movements','Audit Trail ของ Stock']};
const esc=v=>String(v??'').replace(/[&<>"']/g,s=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[s]));const fmt=v=>Number(v||0).toLocaleString(undefined,{maximumFractionDigits:2});const dateFmt=v=>v?new Date(v).toLocaleString():'-';const dFmt=v=>v?new Date(v).toLocaleDateString():'-';
async function api(url,opt={}){opt.headers={...(opt.headers||{}),Authorization:'Bearer '+token};if(opt.body&&!opt.headers['Content-Type']&&!(opt.body instanceof FormData))opt.headers['Content-Type']='application/json';const r=await fetch(url,opt);if(r.status===401){logout();throw new Error(validationMessage('session expired'))}const ct=r.headers.get('content-type')||'';const d=ct.includes('application/json')?await r.json():await r.blob();if(!r.ok)throw new Error(validationMessage(d.error||'request failed'));return d}
async function login(){try{const r=await fetch('/api/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({username:$('loginUser').value,password:$('loginPass').value})});const d=await r.json();if(!r.ok)throw new Error(validationMessage(d.error));token=d.token;me=d.user;localStorage.setItem('easywms_token',token);localStorage.setItem('easywms_user',JSON.stringify(me));showApp();loadDashboard()}catch(e){$('loginMsg').textContent=e.message;$('loginMsg').className='msg error'}}function logout(){token='';me=null;localStorage.removeItem('easywms_token');localStorage.removeItem('easywms_user');$('appView').classList.add('hidden');$('loginView').classList.remove('hidden')}
function showApp(){$('loginView').classList.add('hidden');$('appView').classList.remove('hidden');$('currentUser').textContent=`${me.name} • ${me.role}`;document.querySelectorAll('.admin-only').forEach(x=>x.style.display=me.role==='ADMIN'?'':'none');loadLocations()}
async function loadLocations(){try{const [locations,inventory]=await Promise.all([api('/api/locations'),api('/api/inventory')]);const occupied=new Set(inventory.filter(x=>Number(x.qty)>0).map(x=>x.location?.code));['rLoc','tTo','aLoc'].forEach(id=>{const field=$(id);if(!field)return;const selected=field.value||'A-01-01';const availableForPut=['rLoc','tTo'].includes(id);const options=locations.filter(x=>!availableForPut||!occupied.has(x.code));const select=document.createElement('select');select.id=id;select.innerHTML='<option value="">Select location</option>'+options.map(x=>`<option value="${esc(x.code)}">${esc(x.code)} - ${esc(x.name||'')} (${esc(x.warehouse_code||'-')}/${esc(x.zone_code||'-')})</option>`).join('');select.value=options.some(x=>x.code===selected)?selected:'';field.replaceWith(select)})}catch(e){console.error('Unable to load locations',e)}}
document.querySelectorAll('.nav').forEach(b=>b.onclick=()=>{document.querySelectorAll('.nav').forEach(x=>x.classList.remove('active'));b.classList.add('active');document.querySelectorAll('.page').forEach(x=>x.classList.remove('active'));$(b.dataset.page).classList.add('active');const [a,c]=titles[b.dataset.page];$('pageTitle').textContent=a;$('pageSub').textContent=c;if(b.dataset.page==='transfer')showTransferHistory();if(b.dataset.page==='issue')showIssueHistory();if(b.dataset.page==='receive')showReceiveHistory();if(b.dataset.page==='master')loadMaster();if(b.dataset.page==='inventory')loadInventory();if(b.dataset.page==='lots')loadLots();if(b.dataset.page==='stockcount')showCountHistory();if(b.dataset.page==='adjustment')showAdjustmentHistory();if(b.dataset.page==='movements')loadMovements()});
function table(headers,rows){return `<div class="table-wrap"><table><thead><tr>${headers.map(h=>`<th>${h}</th>`).join('')}</tr></thead><tbody>${rows.length?rows.join(''):`<tr><td colspan="${headers.length}" class="empty">No data</td></tr>`}</tbody></table></div>`}function badge(v){let c=['IN','TRANSFER_IN','ADJUST_IN','APPROVED','COMPLETED'].includes(v)?'in':['PENDING','PENDING_APPROVAL'].includes(v)?'pending':'out';return `<span class="badge ${c}">${esc(v)}</span>`}
const tableSearchStates = new Map();

function validationMessage(message) {
  const text = String(message || 'request failed');
  const translations = {
    'product not found': 'ไม่พบสินค้า',
    'location not found': 'ไม่พบตำแหน่งจัดเก็บ',
    'destination not found': 'ไม่พบตำแหน่งปลายทาง',
    'source and destination must differ': 'ต้นทางและปลายทางต้องไม่เป็นตำแหน่งเดียวกัน',
    'insufficient stock': 'สต็อกคงเหลือไม่เพียงพอ',
    'insufficient lot stock': 'สต็อกใน Lot ไม่เพียงพอ',
    'insufficient stock for adjustment': 'สต็อกคงเหลือไม่เพียงพอสำหรับการปรับยอด',
    'supplier not found': 'ไม่พบผู้ขาย',
    'customer not found': 'ไม่พบลูกค้า',
    'customer is required for customer delivery': 'กรุณาเลือกลูกค้าสำหรับการเบิกส่งลูกค้า',
    'supplier is only allowed for receiving goods': 'ระบุผู้ขายได้เฉพาะการรับสินค้า',
    'customer is only allowed for customer delivery': 'ระบุลูกค้าได้เฉพาะการเบิกส่งลูกค้า',
    'invalid issue purpose': 'วัตถุประสงค์การเบิกสินค้าไม่ถูกต้อง',
    'direction must be IN or OUT': 'ประเภทการปรับยอดต้องเป็น IN หรือ OUT',
    'adjustment is not pending': 'รายการปรับยอดไม่ได้อยู่ในสถานะรออนุมัติ',
    'warehouse, zone, code and name are required': 'กรุณาระบุคลังสินค้า โซน รหัส และชื่อ',
    'warehouse not found': 'ไม่พบคลังสินค้า',
    'zone does not belong to the selected warehouse': 'โซนไม่ได้อยู่ในคลังสินค้าที่เลือก',
    'invalid username or password': 'ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง',
    'permission denied': 'คุณไม่มีสิทธิ์ดำเนินการนี้',
    'unauthorized': 'กรุณาเข้าสู่ระบบ',
    'session expired': 'เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่',
    'request failed': 'ดำเนินการไม่สำเร็จ',
    'file is required': 'กรุณาเลือกไฟล์',
    'invalid csv': 'ไฟล์ CSV ไม่ถูกต้อง',
    'เลือกต้นทางที่มีสินค้าใน Inventory': 'Select a source location with available inventory.',
    'ไม่มีสินค้าคงเหลือให้เบิก กรุณา Refresh เพื่อตรวจสอบสต็อกอีกครั้ง': 'No stock available. Refresh to check inventory again.',
    'กรุณาเลือก Location ที่ต้องการเบิกสินค้า': 'Select the location to issue stock from.',
    'กรุณาเลือกรายการสินค้า / Lot ที่ต้องการเบิก': 'Select the product / lot to issue.',
    'กรุณาระบุจำนวนที่ต้องการเบิกเป็นตัวเลขมากกว่า 0': 'Enter an issue quantity greater than zero.',
    'กรุณาเลือกลูกค้าสำหรับการเบิกส่งลูกค้า': 'Select a customer for customer delivery.',
    'กรุณาเลือก Location ที่ต้องการตรวจนับ': 'Select a location for stock counting.',
    'กรุณาเลือกสินค้าที่ต้องการตรวจนับ': 'Select a product for stock counting.',
    'กรุณากรอกจำนวนที่นับได้จริงตั้งแต่ 0 ขึ้นไป': 'Enter a counted quantity of zero or more.'
  };
  if (translations[text]) return /[\u0e00-\u0e7f]/.test(text) ? `${text}\n${translations[text]}` : `${translations[text]}\n${text}`;
  if (text.includes('\n') && /[\u0e00-\u0e7f]/.test(text)) return text;
  if (text.startsWith('จำนวนเบิกเกินยอดคงเหลือ เบิกได้สูงสุด ')) return `${text}\nInsufficient stock. Maximum issue quantity: ${text.slice('จำนวนเบิกเกินยอดคงเหลือ เบิกได้สูงสุด '.length)}`;
  const failures = [...text.matchAll(/Field validation for '([^']+)' failed on the '([^']+)' tag/g)];
  if (failures.length) return failures.map(([, field, rule]) => rule === 'required'
    ? `กรุณาระบุ ${field}\n${field} is required.`
    : rule === 'gt' ? `${field} ต้องมากกว่า 0\n${field} must be greater than zero.`
    : rule === 'gte' ? `${field} ต้องมีค่าตั้งแต่ 0 ขึ้นไป\n${field} must be zero or greater.`
    : `ข้อมูล ${field} ไม่ถูกต้อง\n${field} is invalid (${rule}).`).join('\n\n');
  return /[\u0e00-\u0e7f]/.test(text) ? `${text}\nUnable to complete the request. Please check the input.` : `ดำเนินการไม่สำเร็จ กรุณาตรวจสอบข้อมูล\n${text}`;
}

function dateSearchValue(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

function renderSearchTable(targetId, markup, options = {}) {
  const target = $(targetId);
  const key = targetId + ':' + (options.key || '');
  const template = document.createElement('template');
  template.innerHTML = markup;
  const wrapper = template.content.firstElementChild;
  const headers = [...wrapper.querySelectorAll('thead th')];
  const dropdownColumns = headers.map(header => ['status', 'สถานะ', 'active', 'type', 'ประเภท', 'role'].includes(header.textContent.trim().toLowerCase()));
  const dateColumns = headers.map(header => /^(date|mfg|exp|หมดอายุ)$|^วันที่/i.test(header.textContent.trim()));
  const body = wrapper.querySelector('tbody');
  const rows = [...body.rows].filter(row => !row.querySelector('.empty'));
  const texts = rows.map(row => [...row.cells].map((cell, index) => dateColumns[index] ? (cell.dataset.searchDate || '') : cell.textContent.trim().toLocaleLowerCase()));
  const filters = tableSearchStates.get(key) || headers.map(() => '');
  tableSearchStates.set(key, filters);
  let page = options.page || 1;
  const filterRow = document.createElement('tr');
  filterRow.className = 'receive-filters';
  headers.forEach((header, index) => {
    const cell = document.createElement('td');
    if (dateColumns[index]) {
      cell.innerHTML = `<input class="table-filter-date" type="date" aria-label="ค้นหา ${esc(header.textContent)}" value="${esc(filters[index])}">`;
      cell.querySelector('input').addEventListener('change', event => {
        filters[index] = event.target.value;
        page = 1;
        update();
      });
    } else if (dropdownColumns[index]) {
      const values = [...new Set(rows.map(row => row.cells[index].textContent.trim()).filter(Boolean))];
      if (filters[index] && !values.includes(filters[index])) values.push(filters[index]);
      values.sort((a, b) => a.localeCompare(b));
      cell.innerHTML = `<select class="table-filter-select" aria-label="ค้นหา ${esc(header.textContent)}"><option value="">ทั้งหมด</option>${values.map(value => `<option value="${esc(value)}">${esc(value)}</option>`).join('')}</select>`;
      const select = cell.querySelector('select');
      select.value = filters[index];
      select.addEventListener('change', event => {
        filters[index] = event.target.value;
        page = 1;
        update();
      });
    } else if (!['Action', 'Label', 'ดำเนินการ', 'เลือก'].includes(header.textContent.trim())) {
      cell.innerHTML = `<div class="receive-search"><input type="search" aria-label="ค้นหา ${esc(header.textContent)}" value="${esc(filters[index])}"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><circle cx="10.5" cy="10.5" r="6.5"></circle><path d="m15.5 15.5 5 5"></path></svg></div>`;
      cell.querySelector('input').addEventListener('input', event => {
        filters[index] = event.target.value;
        page = 1;
        update();
      });
    }
    filterRow.append(cell);
  });
  wrapper.querySelector('thead').append(filterRow);
  const pagination = document.createElement('nav');
  pagination.className = 'master-pagination';
  pagination.setAttribute('aria-label', 'หน้ารายการ');
  pagination.addEventListener('click', event => {
    const button = event.target.closest('button[data-direction]');
    if (!button || button.disabled) return;
    page += Number(button.dataset.direction);
    update();
  });
  target.replaceChildren(wrapper, pagination);

  function update() {
    const terms = filters.map(value => value.trim().toLocaleLowerCase());
    const matched = rows.filter((row, rowIndex) => terms.every((term, column) => !term || (dropdownColumns[column] || dateColumns[column] ? texts[rowIndex][column] === term : (texts[rowIndex][column] || '').includes(term))));
    const size = options.pageSize || Math.max(1, matched.length);
    const pages = Math.max(1, Math.ceil(matched.length / size));
    page = Math.max(1, Math.min(page, pages));
    const start = (page - 1) * size;
    body.replaceChildren(...matched.slice(start, start + size));
    if (!matched.length) body.innerHTML = `<tr><td colspan="${headers.length}" class="empty">${terms.some(Boolean) ? 'ไม่พบรายการที่ตรงกับการค้นหา' : 'ไม่มีข้อมูล'}</td></tr>`;
    pagination.innerHTML = `<span role="status">แสดง ${matched.length ? start + 1 : 0}–${Math.min(start + size, matched.length)} จาก ${matched.length} รายการ</span>${options.pageSize ? `<div><button class="ghost" data-direction="-1" ${page === 1 ? 'disabled' : ''}>ก่อนหน้า</button><span>หน้า ${page} / ${pages}</span><button class="ghost" data-direction="1" ${page === pages ? 'disabled' : ''}>ถัดไป</button></div>` : ''}`;
    options.onPage?.(page);
  }
  update();
}
