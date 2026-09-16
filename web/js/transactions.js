async function postAction(url,body,msgId){const m=$(msgId);try{const d=await api(url,{method:'POST',body:JSON.stringify(body)});m.className='msg success';m.textContent=(d.message||'completed')+' ✓';loadDashboard();return d}catch(e){m.className='msg error';m.textContent=e.message;throw e}}
async function submitMovement(type){if(type==='issue')return submitIssue();if(type==='receive'&&$('receiveSubmit').disabled)return;const p=type==='receive'?'r':'i',body={sku:$(p+'Sku').value,location_code:$(p+'Loc').value,qty:Number($(p+'Qty').value),lot_no:$(p+'Lot').value,reference:$(p+'Ref').value,reason_code:$(p+'Reason').value};if(type==='receive'){body.supplier_code=$('rSupplier').value;if($('rMfg').value)body.mfg_date=new Date($('rMfg').value+'T00:00:00Z').toISOString();if($('rExp').value)body.exp_date=new Date($('rExp').value+'T00:00:00Z').toISOString()}if(type==='receive')$('receiveSubmit').disabled=true;try{await postAction('/api/'+type,body,type+'Msg');if(type==='receive'){clearReceiveForm();showReceiveHistory('รับสินค้าเข้าคลังเรียบร้อยแล้ว');loadLocations()}}catch{}finally{if(type==='receive')$('receiveSubmit').disabled=false}}
async function submitTransfer() {
  if ($('transferSubmit').disabled) return;
  $('transferSubmit').disabled = true;
  try {
    await postAction('/api/transfer',{sku:$('tSku').value,from_location:$('tFrom').value,to_location:$('tTo').value,qty:Number($('tQty').value),reference:$('tRef').value,reason_code:$('tReason').value},'transferMsg');
    await showTransferHistory('ย้ายสินค้าเรียบร้อยแล้ว');
    await loadLocations();
  } catch {} finally { $('transferSubmit').disabled = false; }
}

async function openTransferForm() {
  if ($('transferSubmit').disabled) return;
  ['tSku','tFrom','tTo','tQty','tRef'].forEach(id => $(id).value = '');
  $('tReason').value = 'TRF';
  $('transferMsg').textContent = '';
  $('transferMsg').className = 'msg';
  $('transferHistory').classList.add('hidden');
  $('transferForm').classList.remove('hidden');
  $('pageSub').textContent = 'ย้ายสินค้าระหว่าง Location';
  await loadLocations();
  $('tSku').focus();
}

function showTransferHistory(message = '') {
  $('transferForm').classList.add('hidden');
  $('transferHistory').classList.remove('hidden');
  $('pageSub').textContent = 'ประวัติการย้ายสินค้า';
  $('transferHistoryMsg').className = 'msg success';
  $('transferHistoryMsg').textContent = message;
  return loadTransferHistory();
}

const transferPageSize = 5;
let transferPage = 1, transferRows = [], transferHistoryLoadId = 0;

async function loadTransferHistory() {
  const loadId = ++transferHistoryLoadId;
  transferPage = 1;
  transferRows = [];
  $('transferTable').textContent = 'กำลังโหลดประวัติการย้ายสินค้า...';
  try {
    // Each transfer has two movements; show its outgoing movement once.
    const rows = await api('/api/movements?type=TRANSFER_OUT');
    if (loadId !== transferHistoryLoadId) return;
    transferRows = rows || [];
    renderTransferPage();
  } catch (e) {
    if (loadId !== transferHistoryLoadId) return;
    $('transferTable').textContent = 'ไม่สามารถโหลดประวัติการย้ายสินค้า: ' + e.message;
  }
}

function renderTransferPage() {
  const total = transferRows.length, pages = Math.max(1, Math.ceil(total / transferPageSize));
  transferPage = Math.max(1, Math.min(transferPage, pages));
  const start = (transferPage - 1) * transferPageSize;
  const target = $('transferTable');
  target.innerHTML = table(['วันที่ย้าย','เอกสารอ้างอิง','SKU','สินค้า','ต้นทาง','ปลายทาง','จำนวน','หน่วย','ผู้ย้าย'], transferRows.slice(start, start + transferPageSize).map(x => `<tr><td>${dateFmt(x.created_at)}</td><td>${esc(x.reference || '-')}</td><td>${esc(x.product?.sku || '-')}</td><td>${esc(x.product?.name || '-')}</td><td>${esc(x.location?.code || '-')}</td><td>${esc(x.note?.startsWith('To ') ? x.note.slice(3) : '-')}</td><td>${fmt(x.qty)}</td><td>${esc(x.product?.unit || '-')}</td><td>${esc(x.created_by || '-')}</td></tr>`));
  if (!total) target.querySelector('.empty').textContent = 'ยังไม่มีประวัติการย้ายสินค้า';
  target.insertAdjacentHTML('beforeend', `<nav class="master-pagination" aria-label="หน้าประวัติการย้ายสินค้า"><span role="status">แสดง ${total ? start + 1 : 0}–${Math.min(start + transferPageSize,total)} จาก ${total} รายการ</span><div><button class="ghost" onclick="changeTransferPage(-1)" ${transferPage === 1 ? 'disabled' : ''}>ก่อนหน้า</button><span>หน้า ${transferPage} / ${pages}</span><button class="ghost" onclick="changeTransferPage(1)" ${transferPage === pages ? 'disabled' : ''}>ถัดไป</button></div></nav>`);
}

function changeTransferPage(direction) {
  transferPage += direction;
  renderTransferPage();
}
async function submitStockCount(){try{await postAction('/api/stock-count',{sku:$('cSku').value,location_code:$('cLoc').value,counted_qty:Number($('cQty').value),reference:$('cRef').value,reason_code:$('cReason').value},'stockCountMsg');loadStockCounts()}catch{}}
async function submitAdjustment(){try{await postAction('/api/adjustments',{sku:$('aSku').value,location_code:$('aLoc').value,direction:$('aDirection').value,qty:Number($('aQty').value),reference:$('aRef').value,reason_code:$('aReason').value,note:$('aNote').value},'adjustmentMsg');loadAdjustments()}catch{}}
function clearReceiveForm() {
  $('rSupplier').value = '';
  ['rSku','rLoc','rQty','rLot','rMfg','rExp','rRef'].forEach(id => $(id).value = '');
  $('rReason').value = 'RCV';
  $('receiveMsg').textContent = '';
  $('receiveMsg').className = 'msg';
}

async function openReceiveForm() {
  clearReceiveForm();
  $('receiveHistory').classList.add('hidden');
  $('receiveForm').classList.remove('hidden');
  $('pageSub').textContent = 'รับสินค้าเข้าคลัง';
  await Promise.all([loadLocations(), loadReceiveSuppliers()]);
  $('rSku').focus();
}

async function loadReceiveSuppliers() {
  $('rSupplier').innerHTML = '<option value="">ไม่ระบุผู้ขาย</option>';
  $('rSupplier').disabled = true;
  $('receiveSupplierHint').textContent = 'กำลังโหลดผู้ขาย...';
  try {
    const suppliers = await api('/api/suppliers');
    $('rSupplier').innerHTML += (suppliers || []).map(x => `<option value="${esc(x.code)}">${esc(x.code)} - ${esc(x.name)}</option>`).join('');
    $('rSupplier').disabled = false;
    $('receiveSupplierHint').textContent = suppliers?.length ? '' : 'ยังไม่มีผู้ขาย เพิ่มได้ที่ Master Data > Supplier';
  } catch (e) {
    $('receiveSupplierHint').textContent = 'โหลดผู้ขายไม่สำเร็จ: ' + e.message + ' กรุณาปิดแล้วเปิดฟอร์มใหม่';
  }
}

function showReceiveHistory(message = '') {
  $('receiveForm').classList.add('hidden');
  $('receiveHistory').classList.remove('hidden');
  $('pageSub').textContent = 'ประวัติการรับสินค้า';
  $('receiveHistoryMsg').className = 'msg success';
  $('receiveHistoryMsg').textContent = message;
  return loadReceiveHistory();
}

const receivePageSize = 5;
let receivePage = 1, receiveRows = [], receiveLoadId = 0;

async function loadReceiveHistory() {
  const loadId = ++receiveLoadId;
  const target = $('receiveTable');
  receivePage = 1;
  receiveRows = [];
  target.textContent = 'กำลังโหลดประวัติการรับสินค้า...';
  try {
    const rows = await api('/api/movements?type=IN');
    if (loadId !== receiveLoadId) return;
    receiveRows = rows || [];
    renderReceivePage();
  } catch (e) {
    if (loadId !== receiveLoadId) return;
    target.textContent = 'ไม่สามารถโหลดประวัติการรับสินค้า: ' + e.message;
  }
}

function renderReceivePage() {
  const target = $('receiveTable');
  const total = receiveRows.length;
  const pages = Math.max(1, Math.ceil(total / receivePageSize));
  receivePage = Math.max(1, Math.min(receivePage, pages));
  const start = (receivePage - 1) * receivePageSize;
  const rows = receiveRows.slice(start, start + receivePageSize);
  target.innerHTML = table(
      ['วันที่รับ','เอกสารอ้างอิง','ผู้ขาย','SKU','สินค้า','คลังสินค้า','Location','จำนวน','หน่วย','Lot / Batch','ผู้รับ'],
      rows.map(x => `<tr><td>${dateFmt(x.created_at)}</td><td>${esc(x.reference || '-')}</td><td>${esc(x.supplier_code ? x.supplier_code + ' - ' + x.supplier_name : '-')}</td><td>${esc(x.product?.sku || '-')}</td><td>${esc(x.product?.name || '-')}</td><td>${esc(x.location?.warehouse_code || '-')}</td><td>${esc(x.location?.code || '-')}</td><td>${fmt(x.qty)}</td><td>${esc(x.product?.unit || '-')}</td><td>${esc(x.lot_no || '-')}</td><td>${esc(x.created_by || '-')}</td></tr>`)
    );
  if (!total) target.querySelector('.empty').textContent = 'ยังไม่มีประวัติการรับสินค้า';
  target.insertAdjacentHTML('beforeend', `<nav class="master-pagination" aria-label="หน้าประวัติการรับสินค้า"><span role="status">แสดง ${total ? start + 1 : 0}–${Math.min(start + receivePageSize, total)} จาก ${total} รายการ</span><div><button class="ghost" onclick="changeReceivePage(-1)" ${receivePage === 1 ? 'disabled' : ''}>ก่อนหน้า</button><span>หน้า ${receivePage} / ${pages}</span><button class="ghost" onclick="changeReceivePage(1)" ${receivePage === pages ? 'disabled' : ''}>ถัดไป</button></div></nav>`);
}

function changeReceivePage(direction) {
  receivePage += direction;
  renderReceivePage();
}
let issueInventory = [], issueLots = [], issueLoading = false, issueSaving = false, issueLoadId = 0;

async function loadIssueInventory() {
  const loadId = ++issueLoadId;
  issueLoading = true;
  updateIssueAvailable();
  try {
    const [inventory, lots] = await Promise.all([api('/api/inventory'), api('/api/inventory-lots')]);
    if (loadId !== issueLoadId) return false;
    issueInventory = inventory || [];
    issueLots = lots || [];
    issueLoading = false;
    updateIssueLocations();
    return true;
  } catch (e) {
    if (loadId !== issueLoadId) return false;
    issueInventory = [];
    issueLots = [];
    issueLoading = false;
    updateIssueLocations();
    $('issueMsg').className = 'msg error';
    $('issueMsg').textContent = e.message;
    return false;
  }
}

let issueChoices = [], issueChoiceKey = '';

function updateIssueLocations() {
  const field = $('iLoc'), selected = field.value;
  const locations = [...new Map(issueInventory.filter(x => Number(x.qty) > 0).map(x => [x.location.code, x.location])).values()].sort((a,b) => a.code.localeCompare(b.code));
  field.innerHTML = '<option value="">เลือก Location ที่ต้องการเบิก</option>' + locations.map(x => `<option value="${esc(x.code)}">${esc(x.code)} — ${esc(x.name)} (${esc(x.warehouse_code)})</option>`).join('');
  field.value = locations.some(x => x.code === selected) ? selected : locations.length === 1 ? locations[0].code : '';
  field.disabled = !locations.length || issueLoading || issueSaving;
  updateIssueItems();
}

function updateIssueItems() {
  issueChoices = [];
  for (const stock of issueInventory.filter(x => x.location.code === $('iLoc').value && Number(x.qty) > 0)) {
    const lots = issueLots.filter(x => x.product_id === stock.product_id && x.location_id === stock.location_id && Number(x.qty) > 0).sort((a,b) => (a.exp_date || '9999').localeCompare(b.exp_date || '9999'));
    for (const lot of lots) issueChoices.push({key: lot.id, stock, lot: lot.lot_no, expiry: lot.exp_date, qty: Math.min(Number(stock.qty),Number(lot.qty))});
    const untracked = Number(stock.qty) - lots.reduce((sum,x) => sum + Number(x.qty),0);
    if (untracked > 0) issueChoices.push({key: stock.id, stock, lot: '', expiry: null, qty: untracked});
  }
  if (!issueChoices.some(x => x.key === issueChoiceKey)) issueChoiceKey = issueChoices.length === 1 ? issueChoices[0].key : '';
  $('issueItems').innerHTML = !$('iLoc').value ? '<p>เลือก Location เพื่อดูสินค้าที่เบิกได้</p>' : table(['เลือก','สินค้า','Lot','หมดอายุ','คงเหลือ'], issueChoices.map((x,i) => `<tr><td><input type="radio" name="issueChoice" aria-label="${esc(x.stock.product.sku + ' ' + (x.lot || 'ไม่ระบุ Lot'))}" ${x.key === issueChoiceKey ? 'checked' : ''} onchange="selectIssueItem(${i})" style="width:auto"></td><td>${esc(x.stock.product.sku)}<br>${esc(x.stock.product.name)}</td><td>${esc(x.lot || 'ไม่มี Lot')}</td><td>${dFmt(x.expiry)}</td><td>${fmt(x.qty)} ${esc(x.stock.product.unit)}</td></tr>`));
  updateIssueAvailable();
}

function selectIssueItem(index) {
  issueChoiceKey = issueChoices[index]?.key || '';
  $('iQty').value = '1';
  updateIssueAvailable();
}

function selectedIssueItem() { return issueChoices.find(x => x.key === issueChoiceKey); }

function updateIssueAvailable() {
  const item = selectedIssueItem(), qty = Number($('iQty').value);
  const valid = item && Number.isFinite(qty) && qty > 0 && qty <= item.qty;
  $('iQty').max = String(item?.qty || 0);
  $('iQty').disabled = !item || issueLoading || issueSaving;
  $('iLoc').disabled = issueLoading || issueSaving || !issueInventory.some(x => Number(x.qty)>0);
  document.querySelectorAll('[name="issueChoice"]').forEach(x => x.disabled = issueLoading || issueSaving);
  $('issueSubmit').disabled = issueLoading || issueSaving;
  $('issueSubmit').textContent = issueSaving ? 'กำลังบันทึก...' : issueLoading ? 'กำลังโหลดสต็อก...' : 'ยืนยันเบิกสินค้า';
  $('issueAvailable').textContent = issueLoading ? 'กำลังโหลด Inventory...' : !issueInventory.some(x => Number(x.qty)>0) ? 'ไม่มีสินค้าคงเหลือให้เบิก' : !item ? 'เลือก Location และรายการสินค้าเพื่อเบิก' : `เบิกได้สูงสุด ${fmt(item.qty)} ${item.stock.product.unit}${valid ? '' : ' — ระบุจำนวนมากกว่า 0 และไม่เกินยอดคงเหลือ'}`;
  $('issueSummary').textContent = item ? `${item.stock.product.name} • ${item.stock.location.code} • Lot: ${item.lot || '-'} • จำนวน ${fmt(qty)} ${item.stock.product.unit}` : '';
  if (item) $('issueSummary').textContent += $('iPurpose').value === 'CUSTOMER' ? ' • ลูกค้า: ' + ($('iCustomer').value ? $('iCustomer').selectedOptions[0].textContent : 'ยังไม่ได้เลือก') : ' • เบิกใช้ภายใน';
}

async function submitIssue() {
  if (issueSaving || issueLoading) return;
  const item = selectedIssueItem(), qty = Number($('iQty').value);
  let error = '', field = $('iLoc');
  if (!issueInventory.some(x => Number(x.qty) > 0)) error = 'ไม่มีสินค้าคงเหลือให้เบิก กรุณา Refresh เพื่อตรวจสอบสต็อกอีกครั้ง';
  else if (!$('iLoc').value) error = 'กรุณาเลือก Location ที่ต้องการเบิกสินค้า';
  else if (!item) {
    error = 'กรุณาเลือกรายการสินค้า / Lot ที่ต้องการเบิก';
    field = document.querySelector('[name="issueChoice"]') || $('iLoc');
  } else if (!$('iQty').value.trim() || !Number.isFinite(qty) || qty <= 0) {
    error = 'กรุณาระบุจำนวนที่ต้องการเบิกเป็นตัวเลขมากกว่า 0';
    field = $('iQty');
  } else if (qty > item.qty) {
    error = `จำนวนเบิกเกินยอดคงเหลือ เบิกได้สูงสุด ${fmt(item.qty)} ${item.stock.product.unit}`;
    field = $('iQty');
  }
  if (!error && $('iPurpose').value === 'CUSTOMER' && !$('iCustomer').value) {
    error = 'กรุณาเลือกลูกค้าสำหรับการเบิกส่งลูกค้า';
    field = $('iCustomer');
  }
  if (error) {
    $('issueMsg').className = 'msg error';
    $('issueMsg').textContent = error;
    alert(error);
    field.focus();
    return;
  }
  $('issueMsg').textContent = '';
  if (!confirm('ยืนยันเบิกสินค้า?\n' + $('issueSummary').textContent)) return;
  const body = {sku: item.stock.product.sku, location_code: item.stock.location.code, qty, lot_no: item.lot, reference: $('iRef').value.trim(), reason_code: 'ISS', note: $('iNote').value.trim()};
  body.issue_purpose = $('iPurpose').value;
  body.customer_code = body.issue_purpose === 'CUSTOMER' ? $('iCustomer').value : '';
  issueSaving = true;
  updateIssueAvailable();
  try {
    await postAction('/api/issue', body, 'issueMsg');
    await showIssueHistory('เบิกสินค้าเรียบร้อยแล้ว');
    $('iQty').value = '1';
    await loadIssueInventory();
  } catch (e) {
    alert('ไม่สามารถเบิกสินค้าได้: ' + e.message);
    await loadIssueInventory();
  } finally {
    issueSaving = false;
    updateIssueAvailable();
  }
}

const issuePageSize = 5;
let issuePage = 1, issueRows = [], issueHistoryLoadId = 0;

async function loadIssueHistory() {
  const loadId = ++issueHistoryLoadId;
  const target = $('issueTable');
  issuePage = 1;
  issueRows = [];
  target.textContent = 'กำลังโหลดประวัติการเบิกสินค้า...';
  try {
    const rows = await api('/api/movements?type=OUT');
    if (loadId !== issueHistoryLoadId) return;
    issueRows = rows || [];
    renderIssuePage();
  } catch (e) {
    if (loadId !== issueHistoryLoadId) return;
    target.textContent = 'ไม่สามารถโหลดประวัติการเบิกสินค้า: ' + e.message;
  }
}

function renderIssuePage() {
  const target = $('issueTable');
  const total = issueRows.length;
  const pages = Math.max(1, Math.ceil(total / issuePageSize));
  issuePage = Math.max(1, Math.min(issuePage, pages));
  const start = (issuePage - 1) * issuePageSize;
  const rows = issueRows.slice(start, start + issuePageSize);
  target.innerHTML = table(
      ['วันที่เบิก','เอกสารอ้างอิง','ลูกค้า','SKU','สินค้า','คลังสินค้า','Location','จำนวน','หน่วย','Lot / Batch','ผู้เบิก'],
      rows.map(x => `<tr><td>${dateFmt(x.created_at)}</td><td>${esc(x.reference || '-')}</td><td>${esc(x.customer_code ? x.customer_code + ' - ' + x.customer_name : x.issue_purpose === 'INTERNAL' ? 'เบิกใช้ภายใน' : '-')}</td><td>${esc(x.product?.sku || '-')}</td><td>${esc(x.product?.name || '-')}</td><td>${esc(x.location?.warehouse_code || '-')}</td><td>${esc(x.location?.code || '-')}</td><td>${fmt(x.qty)}</td><td>${esc(x.product?.unit || '-')}</td><td>${esc(x.lot_no || '-')}</td><td>${esc(x.created_by || '-')}</td></tr>`)
    );
  if (!total) target.querySelector('.empty').textContent = 'ยังไม่มีประวัติการเบิกสินค้า';
  target.insertAdjacentHTML('beforeend', `<nav class="master-pagination" aria-label="หน้าประวัติการเบิกสินค้า"><span role="status">แสดง ${total ? start + 1 : 0}–${Math.min(start + issuePageSize, total)} จาก ${total} รายการ</span><div><button class="ghost" onclick="changeIssuePage(-1)" ${issuePage === 1 ? 'disabled' : ''}>ก่อนหน้า</button><span>หน้า ${issuePage} / ${pages}</span><button class="ghost" onclick="changeIssuePage(1)" ${issuePage === pages ? 'disabled' : ''}>ถัดไป</button></div></nav>`);
}

function changeIssuePage(direction) {
  issuePage += direction;
  renderIssuePage();
}
async function openIssueForm() {
  if (issueSaving) return;
  issueChoiceKey = '';
  issueChoices = [];
  $('iLoc').value = '';
  $('iQty').value = '1';
  $('iRef').value = '';
  $('iNote').value = '';
  $('iPurpose').value = 'INTERNAL';
  $('iCustomer').value = '';
  updateIssueCustomer();
  $('issueMsg').textContent = '';
  $('issueMsg').className = 'msg';
  $('issueForm').querySelector('details').open = false;
  $('issueHistory').classList.add('hidden');
  $('issueForm').classList.remove('hidden');
  $('pageSub').textContent = 'เบิกสินค้าออก';
  await Promise.all([loadIssueInventory(), loadIssueCustomers()]);
  $('iLoc').focus();
}

function updateIssueCustomer() {
  const required = $('iPurpose').value === 'CUSTOMER';
  $('issueCustomerField').classList.toggle('hidden', !required);
  $('iCustomer').required = required;
  if (!required) $('iCustomer').value = '';
  updateIssueAvailable();
}

async function loadIssueCustomers() {
  $('iCustomer').innerHTML = '<option value="">เลือกลูกค้า</option>';
  $('iCustomer').disabled = true;
  $('issueCustomerHint').textContent = 'กำลังโหลดลูกค้า...';
  try {
    const customers = await api('/api/customers');
    $('iCustomer').innerHTML += (customers || []).map(x => `<option value="${esc(x.code)}">${esc(x.code)} - ${esc(x.name)}</option>`).join('');
    $('iCustomer').disabled = false;
    $('issueCustomerHint').textContent = customers?.length ? '' : 'ยังไม่มีลูกค้า กรุณาเพิ่มที่ Master Data > Customer';
  } catch (e) {
    $('issueCustomerHint').textContent = 'โหลดลูกค้าไม่สำเร็จ: ' + e.message + ' กรุณาปิดแล้วเปิดฟอร์มใหม่';
  }
}

function showIssueHistory(message = '') {
  $('issueForm').classList.add('hidden');
  $('issueHistory').classList.remove('hidden');
  $('pageSub').textContent = 'ประวัติการเบิกสินค้า';
  $('issueHistoryMsg').className = 'msg success';
  $('issueHistoryMsg').textContent = message;
  return loadIssueHistory();
}
