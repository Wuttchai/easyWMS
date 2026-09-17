async function loadDashboard(){try{const d=await api('/api/dashboard');$('kpiProducts').textContent=d.products;$('kpiQty').textContent=fmt(d.total_qty);$('kpiWarehouses').textContent=d.warehouses;$('kpiLocations').textContent=d.locations;$('kpiEmployees').textContent=d.employees;$('kpiLow').textContent=d.low_stock;$('kpiPending').textContent=d.pending_adjustments;renderSearchTable('recentTable', table(['Date','Type','SKU','Location','Qty','Lot','User'],(d.recent_movements||[]).map(x=>`<tr><td data-search-date="${dateSearchValue(x.created_at)}">${dateFmt(x.created_at)}</td><td>${badge(x.type)}</td><td>${esc(x.product?.sku)}</td><td>${esc(x.location?.code)}</td><td>${fmt(x.qty)}</td><td>${esc(x.lot_no||'-')}</td><td>${esc(x.created_by||'-')}</td></tr>`)))}catch(e){console.error(e)}}
async function loadInventory(){const d=await api('/api/inventory');renderSearchTable('inventoryTable', table(['SKU','Product','Warehouse','Zone','Location','Qty','Unit','Status'],d.map(x=>`<tr><td><b>${esc(x.product.sku)}</b></td><td>${esc(x.product.name)}</td><td>${esc(x.location.warehouse_code||'-')}</td><td>${esc(x.location.zone_code||'-')}</td><td>${esc(x.location.code)}</td><td>${fmt(x.qty)}</td><td>${esc(x.product.unit)}</td><td>${x.qty<=x.product.min_stock?badge('LOW'):badge('OK')}</td></tr>`)))}
async function loadLots(){const d=await api('/api/inventory-lots');renderSearchTable('lotsTable', table(['SKU','Product','Location','Lot/Batch','MFG','EXP','Qty'],d.map(x=>`<tr><td>${esc(x.product.sku)}</td><td>${esc(x.product.name)}</td><td>${esc(x.location.code)}</td><td><b>${esc(x.lot_no)}</b></td><td data-search-date="${dateSearchValue(x.mfg_date)}">${dFmt(x.mfg_date)}</td><td data-search-date="${dateSearchValue(x.exp_date)}">${dFmt(x.exp_date)}</td><td>${fmt(x.qty)}</td></tr>`)))}
async function loadMovements(){const d=await api('/api/movements');renderSearchTable('movementsTable', table(['Date','Type','SKU','Location','Qty','Lot','Reason','Reference','User'],d.map(x=>`<tr><td data-search-date="${dateSearchValue(x.created_at)}">${dateFmt(x.created_at)}</td><td>${badge(x.type)}</td><td>${esc(x.product.sku)}</td><td>${esc(x.location.code)}</td><td>${fmt(x.qty)}</td><td>${esc(x.lot_no||'-')}</td><td>${esc(x.reason_code||'-')}</td><td>${esc(x.reference||'-')}</td><td>${esc(x.created_by||'-')}</td></tr>`)))}
async function adjAction(id,action){try{await api(`/api/adjustments/${id}/${action}`,{method:'POST'});loadAdjustments(adjustmentPage);loadDashboard();loadInventory()}catch(e){alert(e.message)}}
function exportInventory() { return exportExcel('inventory', 'exportInventoryBtn'); }
function exportMovements() { return exportExcel('movements', 'exportMovementsBtn'); }
async function exportExcel(resource, buttonId) {
  const button = $(buttonId);
  if (button.disabled) return;
  button.disabled = true;
  try {
    const blob = await api('/api/' + resource + '/export');
    const mime = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
    if (!(blob instanceof Blob) || blob.type.split(';')[0] !== mime) {
      throw new Error('เซิร์ฟเวอร์ไม่ได้ส่งไฟล์ Excel กรุณารีสตาร์ตเซิร์ฟเวอร์เวอร์ชันล่าสุดแล้วลองใหม่\nThe server did not return an Excel file. Restart the updated server and try again.');
    }
    const signature = new Uint8Array(await blob.slice(0, 4).arrayBuffer());
    if (signature.length !== 4 || signature[0] !== 0x50 || signature[1] !== 0x4b || signature[2] !== 3 || signature[3] !== 4) {
      throw new Error('ไฟล์ Excel ที่ได้รับไม่ถูกต้อง กรุณาลองส่งออกใหม่\nThe returned Excel file is invalid. Please export again.');
    }
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    const now = new Date();
    const pad = value => String(value).padStart(2, '0');
    const timestamp = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}_${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
    link.download = `${resource}_${timestamp}.xlsx`;
    document.body.append(link);
    link.click();
    link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  } catch (e) {
    alert(validationMessage(e.message));
  } finally { button.disabled = false; }
}
