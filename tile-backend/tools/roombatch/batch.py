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
import argparse, collections, itertools, json, os, sys, urllib.request, importlib.util

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

# --- layout diversity (ORT-132) -------------------------------------------
# A family is one (stageType, openDoors) cell. Rooms in the same family share a
# candidate pool, so two similar ones read as a repeat when the level generator
# puts them next to each other. Measured over the d5 families already shipped:
# ground similarity runs 0.34-1.00 (peak's two rooms have *identical* ground),
# static 0.07-0.50, enemies 0.11-0.43. The ticket asks for difference in "走法
# 和掩体位置" - pathing and cover - so ground and static carry the thresholds.
DIVERSITY = {'ground': 0.95, 'static': 0.35, 'enemy': 0.35}
ENEMY_LAYERS = ('chaser', 'zoner', 'dps', 'mobAir')

def _cells(payload, layer):
    g = payload.get(layer) or []
    return {(y, x) for y, row in enumerate(g) for x, v in enumerate(row) if v}

def _jaccard(a, b):
    return len(a & b) / len(a | b) if (a | b) else 1.0

def signature(payload):
    return {'ground': _cells(payload, 'ground'),
            'static': _cells(payload, 'static'),
            'enemy': set().union(*[_cells(payload, k) for k in ENEMY_LAYERS])}

def family_signatures(stage, ozx_mask):
    """Every template already in this (stage, openDoors) family, as signatures."""
    out = []
    for item in api('/templates?limit=2000')['items']:
        if item.get('stage_type') != stage:
            continue
        if item.get('open_doors') != data_mask(ozx_mask):
            continue
        full = api('/templates/' + item['id'])
        out.append((os.path.basename(item.get('path') or item['id']), signature(full.get('payload') or full)))
    return out

def too_similar(candidate, family):
    """Returns the offending (name, layer, score) or None."""
    for name, sig in family:
        for layer, limit in DIVERSITY.items():
            score = _jaccard(candidate[layer], sig[layer])
            if score >= limit:
                return (name, layer, score)
    return None

def data_mask(ozx_mask):
    m = 0
    for bit in (1, 2, 4, 8):
        if ozx_mask & bit:
            m |= BITS[OZX_TO_DATA[bit]]
    return m

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
    ap.add_argument('--diverse', action='store_true',
                    help='also require each room to differ in layout from every template '
                         'already in its (stage, openDoors) family, and from the others '
                         'produced in this run (ORT-132)')
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
    produced = {}          # (stage, mask) -> signatures made in this run
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
            family = family_signatures(stage, mask) if a.diverse else []
            family += produced.get((stage, mask), [])
            best = None
            for attempt in range(1, a.attempts + 1):
                r = generate(stage, mask, w, h, a.category, shape, ref.get('doors'))
                res = judge.analyse(r)
                if a.diverse:
                    clash = too_similar(signature(r['payload']), family)
                    if clash:
                        res = dict(res)
                        res['fails'] = res['fails'] + [
                            f"too similar to {clash[0]} ({clash[1]} {clash[2]:.2f})"]
                if best is None or score(res) < score(best[2]):
                    best = (attempt, r, res)
                if score(res) == (0, 0, 0):
                    break
            attempt, r, res = best
            ok = not res['fails']
            produced.setdefault((stage, mask), []).append(
                (f"(this run #{len(produced.get((stage, mask), [])) + 1})", signature(r['payload'])))
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

    # The per-room diversity check is incremental: a room is compared against the
    # family as it stood when that room was generated, so a later sibling can end
    # up too close to an earlier one. Re-check every touched family as a whole and
    # name the offending pairs; redo them (delete + rerun that cell) rather than
    # leaving a near-duplicate in a pool the level generator draws from.
    if a.diverse and a.save:
        print("\nfamily re-check (whole-family pairwise):")
        for stage, mask in sorted({k for k in produced}):
            fam = family_signatures(stage, mask)
            worst = []
            for (n1, s1), (n2, s2) in itertools.combinations(fam, 2):
                for layer, limit in DIVERSITY.items():
                    sc = _jaccard(s1[layer], s2[layer])
                    if sc >= limit:
                        worst.append(f"{n1} vs {n2}: {layer} {sc:.2f} (limit {limit})")
            label = f"{stage}/ozx{mask}"
            print(f"  {label:18} pool={len(fam)}  " + ("clean" if not worst else "; ".join(worst)))

main()
