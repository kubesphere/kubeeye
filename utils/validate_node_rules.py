import os
import yaml
from utils.command_security import check_command_security

RULES_DIR = os.path.join(os.path.dirname(__file__), '../rules/node')

results = []

for fname in os.listdir(RULES_DIR):
    if not fname.endswith('.yaml'):
        continue
    path = os.path.join(RULES_DIR, fname)
    with open(path, 'r', encoding='utf-8') as f:
        rule = yaml.safe_load(f)
    # 兼容不同格式
    try:
        cmd = rule['config']['execution']['command']
    except Exception:
        continue
    is_safe, risk, desc = check_command_security(cmd)
    results.append({
        'file': fname,
        'command': cmd,
        'is_safe': is_safe,
        'risk': str(risk),
        'desc': desc
    })

print('Node规则命令安全性校验结果:')
for r in results:
    print(f"{r['file']}: {r['command']} => {'安全' if r['is_safe'] else '禁止'} | 风险: {r['risk']} | {r['desc']}")
