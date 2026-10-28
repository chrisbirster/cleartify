import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
test.beforeEach(async({page})=>{page.on('pageerror', e=>console.log('BROWSER_PAGE_ERROR: '+e.message));page.on('console',m=>{if(m.type()==='error')console.log('BROWSER_CONSOLE_ERROR: '+m.text())});});
const base = 'http://127.0.0.1:3000';
const logPath = '.local/browser-backend.log';
const phone = '+15550002222';
test('local B authenticates, installs A synthetic guide, creates follow-up, gets in-app nudge, completes', async ({ page }) => {
  await page.goto(base + '/clear');
  console.log('CLEAR_DOCUMENT: '+(await page.locator('body').innerText()).slice(0,500));
  await expect(page.getByRole('heading', { name: /Text Clear/ })).toBeVisible();
  await page.getByPlaceholder('+15551234567').fill(phone);
  const loginInput=page.getByPlaceholder('+15551234567');
  console.log('LOGIN_FIELD_VALIDITY: '+JSON.stringify(await loginInput.evaluate((el:HTMLInputElement)=>({pattern:el.pattern,valid:el.checkValidity(),patternMismatch:el.validity.patternMismatch}))));
  await page.getByRole('button', {name:'Request sign-in code'}).click();
  console.log('LOGIN_AFTER_CLICK: '+(await page.locator('body').innerText()).slice(-800));
  await expect(page.getByText(/If the phone belongs to an active account/)).toBeVisible();
  let code = '';
  await expect.poll(() => {
    const log = readFileSync(logPath,'utf8');
    const found = [...log.matchAll(/Your Cleartify sign-in code is (\d{5})\./g)];
    code = found.at(-1)?.[1] ?? '';
    return code.length;
  }, {timeout: 15000}).toBe(5);
  await page.getByPlaceholder('Code from Debug Sender').fill(code);
  await page.getByRole('button',{name:'Verify code'}).click();
  await expect(page.getByRole('heading',{name:'Your Clear conversation'})).toBeVisible();
  await page.getByRole('button',{name:'Refresh published recipes'}).click();
  const recipe = page.getByRole('listitem').filter({hasText:'Local demo: one small step'});
  await expect(recipe).toBeVisible();
  await recipe.getByRole('button',{name:'Preview install'}).click();
  await expect(page.getByText(/Manifest SHA-256:/)).toBeVisible();
  await page.getByLabel(/Also create follow-ups/).check();
  await page.getByRole('button',{name:'Acknowledge manifest and install'}).click();
  await expect(page.getByText(/Installed\. Follow-ups created: 1/)).toBeVisible();
  const first = page.getByRole('listitem').filter({hasText:'Take one small step'});
  await expect(first.first()).toBeVisible();
  // The worker polls at 30s. The fixture guide due time is in the past.
  const nudge = page.getByRole('listitem').filter({hasText:'Take one small step'}).filter({hasText:'in_app'});
  await expect.poll(async () => {
    await page.getByRole('button', {name:'Refresh', exact:true}).click();
    return nudge.count();
  }, {timeout: 60000, intervals:[1000,2500,5000]}).toBeGreaterThan(0);
  await page.getByLabel('Message to Clear').fill('remind me to check my list');
  await page.getByRole('button',{name:'Send command'}).click();
  await expect(page.getByText(/Follow-up scheduled for/)).toBeVisible();
  await page.getByRole('listitem').filter({hasText:'Take one small step'}).getByRole('button',{name:/Complete via command/}).first().click();
  await expect(page.getByLabel('Message to Clear')).toHaveValue(/^done /);
  await page.getByRole('button',{name:'Send command'}).click();
  await expect(page.getByText('Marked done.')).toBeVisible();
  await page.getByRole('button', {name:'Refresh', exact:true}).click();
  await expect(nudge).toHaveCount(0);
});

for (const [name, width, height] of [['desktop',1365,900],['mobile',390,844]] as const) {
 test('screenshots '+name+' of real Clear and recipes; creator is unauthenticated',async({page})=>{
   await page.setViewportSize({width,height});
   for(const route of ['/clear','/recipes','/creator'] as const){
     await page.goto(base+route);
     console.log('SCREEN_DOCUMENT '+route+': '+(await page.locator('body').innerText()).slice(0,200));
     await expect(page.locator('main')).toBeVisible();
     if(route==='/recipes') {
       await expect(page.getByText('Local demo: one small step')).toBeVisible();
       await expect(page.getByText('The catalog returned an unsupported shape. No entries were fabricated.')).toHaveCount(0);
     }
     await page.screenshot({path:'artifacts/screenshots/'+route.slice(1)+'-'+name+'.png',fullPage:true});
   }
 });
}
