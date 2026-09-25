#!/usr/bin/env python3
"""Batch room generation with the agreed acceptance rules.

For every stage x doors category that holds fewer than --target templates, keep
generating until a room passes the HARD rules (see judge.py) or --attempts runs
out, in which case the least-bad attempt is kept. Placement shortfalls are
best-effort and never a rejection.

usage:
  batch.py                         # dry run: show the plan and the verdicts
  batch.py --save                  # same, but write the accepted rooms
  batch.py --target 3 --attempts 30 --stages teaching,building
  batch.py --size 24x10 --category test
"""
import argparse, collections, json, os, sys, urllib.request, importlib.util

SP = os.path.dirname(os.path.abspath(__file__))
API = os.environ.get("ROOM_API", "http://localhost:8099/api/v1")
BITS = {'top': 1, 'right': 2, 'bottom': 4, 'left': 8}
# A filename's openDoors mask is in OZX space, the generator's doors are in data
# space, and the two differ by the ORT-111 rotation: OZX Top(1) is the data right
# edge, Right(2) the data bottom, Bottom(4) the data left, Left(8) the data top.
OZX_TO_DATA = {1: 'right', 2: 'bottom', 4: 'left', 8: 'top'}
SIDES = ['top', 'right', 'bottom', 'left']
STAGE_ORDER = ['start', 'teaching', 'building', 'pressure', 'peak', 'release', 'boss']
ENDPOINT = {'full': 'fullroom', 'bridge': 'bridge', 'platform': 'platform'}

_argv, sys.argv = sys.argv, ['judge', '__skip__']
spec = importlib.util.spec_from_file_location("judge", os.path.join(SP, "judge.py"))
judge = importlib.util.module_from_spec(spec)
try: spec.loader.exec_module(judge)
except Exception: pass
sys.argv = _argv

def api(path, body=None):
    req = urllib.request.Request(API + path,
                                 data=None if body is None else json.dumps(body).encode(),
                                 headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(req, timeout=60) as r:
        return json.loads(r.read())

def dlabel(mask):
    return ''.join(s[0].upper() for s in SIDES if mask & BITS[s]) or '-'

def plan(target, stages):
    items = api('/templates?limit=2000')['items']
    g = collections.defaultdict(list)
    for it in items:
        g[(it.get('stage_type') or '', it.get('open_doors') or 0)].append(it)
    out = []
    for (stage, mask), lst in g.items():
        if not stage or stage in ('start', 'boss'):   # no enemies to judge
            continue
        if stages and stage not in stages:
            continue
        if len(lst) >= target:
            continue
        out.append(((stage, mask), target - len(lst), lst[0]))
    out.sort(key=lambda e: (STAGE_ORDER.index(e[0][0]) if e[0][0] in STAGE_ORDER else 9, e[0][1]))
    return out

def generate(stage, mask, w, h, cat, shape, doors=None):
    body = {'width': w, 'height': h,
            'doors': doors if doors else [s for s in SIDES if mask & BITS[s]],
            'stageType': stage, 'roomCategory': cat, 'softEdgeCount': 3, 'railEnabled': True}
    r = api('/generate/' + ENDPOINT.get(shape, 'fullroom'), body)
    r['_request'] = body
    r['_ref'] = f"{stage}_{mask}"
    return r

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--target', type=int, default=2, help='templates wanted per category')
    ap.add_argument('--attempts', type=int, default=20, help='regeneration attempts per room')
    ap.add_argument('--stages', default='', help='comma-separated stage filter')
    ap.add_argument('--size', default='16x10',
                    help="WxH in data space, default 16x10 (the shipped room size); "
                         "pass 'ref' to follow each category's existing template instead")
    ap.add_argument('--category', default='test', help='roomCategory / subfolder')
    ap.add_argument('--save', action='store_true', help='write accepted rooms (default: dry run)')
    ap.add_argument('--cells', default='',
                    help="explicit work list instead of the top-up plan: "
                         "'stage:ozxMask:count,...' e.g. 'pressure:5:1,boss:4:2'. "
                         "Masks are OZX-space, the way they appear in filenames.")
    a = ap.parse_args()
    stages = [s for s in a.stages.split(',') if s]

    if a.cells:
        todo = []
        for spec in a.cells.split(','):
            stage, mask, n = spec.split(':')
            ozx = int(mask)
            todo.append(((stage, ozx), int(n),
                         {'width': 16, 'height': 10, 'room_type': 'full',
                          'doors': [OZX_TO_DATA[b] for b in (1, 2, 4, 8) if ozx & b]}))
    else:
        todo = plan(a.target, stages)
    if not todo:
        print(f"nothing to do: every category already holds >= {a.target} templates")
        return
    print(f"{'category':18} {'size':7} {'try':>4}  verdict")
    accepted = rejected = 0
    for (stage, mask), n, ref in todo:
        for _ in range(n):
            if a.size == 'ref':
                w, h = ref['width'], ref['height']
            else:
                w, h = (int(v) for v in a.size.lower().split('x'))
            shape = ref.get('room_type') or 'full'
            # Every attempt is generated and the best one kept, rather than
            # taking the first that clears the hard rules: among clean rooms the
            # one that placed the most enemies is the better room.
            def score(res):
                return (len(res['fails']), res['missing'], len(res['soft']))
            best = None
            for attempt in range(1, a.attempts + 1):
                r = generate(stage, mask, w, h, a.category, shape, ref.get('doors'))
                res = judge.analyse(r)
                if best is None or score(res) < score(best[2]):
                    best = (attempt, r, res)
                if score(res) == (0, 0, 0):
                    break
            attempt, r, res = best
            ok = not res['fails']
            accepted += ok; rejected += not ok
            label = f"{stage}_ozx{mask}" if a.cells else f"{stage}_{dlabel(mask)}"
            line = (f"{label:18} {f'{w}x{h}':7} {attempt:>4}  "
                    + ("ACCEPT" if ok else "KEPT-BEST: " + '; '.join(res['fails'])))
            if res['soft']: line += "  | " + '; '.join(res['soft'])
            if a.save:
                p = r['payload']
                p['roomCategory'] = a.category
                created = api('/templates', {'name': f"{stage}-{mask}", 'payload': p})
                path = created.get('path')
                if not path:            # create echoes the model, not the file
                    path = api('/templates/' + created['id']).get('path') or ''
                line += f"\n{'':18} -> {created.get('id')}  {os.path.basename(path)}"
            print(line)
    print(f"\n{accepted} accepted, {rejected} kept-best (hard rules unmet after {a.attempts} attempts)"
          + ("" if a.save else "   [dry run - nothing written]"))

main()
