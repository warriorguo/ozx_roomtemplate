#!/usr/bin/env python3
"""Score generated rooms for 'tidiness'. Everything is computed in VISUAL space
(the data grid rotated 90 CCW) so every number describes the picture the editor
draws and the game renders."""
import json, os, collections, sys

SP = os.path.dirname(os.path.abspath(__file__))
LAYERS = ['ground','softEdge','bridge','rail','mainPath','static','chaser','zoner','dps','mobAir']
SYM = [('rail','R'),('static','S'),('chaser','C'),('zoner','Z'),('dps','D'),
       ('mobAir','A'),('softEdge','E'),('bridge','B')]

def rot(g):
    if not g: return g
    h, w = len(g), len(g[0])
    return [[g[j][w-1-i] for j in range(h)] for i in range(w)]

def comps(g, want=1, diag=False):
    h, w = len(g), len(g[0]); seen=[[0]*w for _ in range(h)]; out=[]
    offs=[(-1,0),(1,0),(0,-1),(0,1)]+([(-1,-1),(-1,1),(1,-1),(1,1)] if diag else [])
    for y in range(h):
        for x in range(w):
            if g[y][x]!=want or seen[y][x]: continue
            st=[(y,x)]; seen[y][x]=1; cells=[]
            while st:
                cy,cx=st.pop(); cells.append((cy,cx))
                for dy,dx in offs:
                    ny,nx=cy+dy,cx+dx
                    if 0<=ny<h and 0<=nx<w and not seen[ny][nx] and g[ny][nx]==want:
                        seen[ny][nx]=1; st.append((ny,nx))
            out.append(cells)
    return out

def quad(cells_list, h, w):
    q=collections.Counter()
    for cells in cells_list:
        cy=sum(c[0] for c in cells)/len(cells); cx=sum(c[1] for c in cells)/len(cells)
        q[('T' if cy<h/2 else 'B')+('L' if cx<w/2 else 'R')]+=1
    return {k:q.get(k,0) for k in ('TL','TR','BL','BR')}

def render(v):
    ground=v['ground']; path=v.get('mainPath'); h,w=len(ground),len(ground[0])
    lines=[]
    for y in range(h):
        row=''
        for x in range(w):
            ch=None
            for k,s in SYM:
                if v.get(k) and v[k][y][x]: ch=s; break
            if ch is None:
                ch=('+' if (path and path[y][x]) else '#') if ground[y][x] else '.'
            row+=ch
        lines.append(row)
    return lines

# Stages that should show holes in the floor rather than a solid slab
# (play-test review, 2026-09-14). The generator raises its carve probability for
# these, but carving is a dice roll, so the rule is enforced here too.
AIRY_STAGES = ('teaching', 'building', 'release')
MAX_AIRY_FILL = 0.9

# The grid layers are always data-space, but `doors` is NOT: fsstore rotates the
# door metadata at the disk boundary (ORT-112/113), so a payload read straight
# from a .json carries OZX-space side names while one from the generate API or
# GET /templates/{id} carries data-space ones. The door-passage rule picks which
# wall to check by name, so it needs to know. Pass _doors_space='ozx' when the
# payload came off disk; the default assumes the API shape.
OZX_TO_DATA_SIDE = {'top': 'right', 'right': 'bottom', 'bottom': 'left', 'left': 'top'}

def data_space_doors(payload, space='data'):
    doors = payload.get('doors') or {}
    if space == 'ozx':
        return {OZX_TO_DATA_SIDE[k]: v for k, v in doors.items() if k in OZX_TO_DATA_SIDE}
    return doors

def analyse(r):
    p=r['payload']
    v={k:rot(p[k]) for k in LAYERS if p.get(k)}
    g=v['ground']; h,w=len(g),len(g[0]); total=h*w
    fill=sum(map(sum,g))/total

    voids=comps(g,0)
    vrect=[]
    for cells in sorted(voids,key=len,reverse=True):
        ys=[c[0] for c in cells]; xs=[c[1] for c in cells]
        area=(max(ys)-min(ys)+1)*(max(xs)-min(xs)+1)
        vrect.append(round(len(cells)/area,2))

    def nb(y,x,val):
        return sum(1 for dy,dx in ((-1,0),(1,0),(0,-1),(0,1))
                   if (g[y+dy][x+dx] if 0<=y+dy<h and 0<=x+dx<w else 0)==val)
    spurs=sum(1 for y in range(h) for x in range(w) if g[y][x] and nb(y,x,0)>=3)
    pits=sum(1 for y in range(h) for x in range(w) if not g[y][x] and nb(y,x,1)>=3)
    sym_lr=sum(1 for y in range(h) for x in range(w) if g[y][x]==g[y][w-1-x])/total
    sym_ud=sum(1 for y in range(h) for x in range(w) if g[y][x]==g[h-1-y][x])/total

    # ground islands: a slab of walkable ground the main body cannot reach.
    gc=sorted(comps(g,1),key=len,reverse=True)
    islands=[c for c in gc[1:]]
    island_cells={c for cells in islands for c in cells}
    island_occupied=0
    for k in ('static','chaser','zoner','dps'):
        if v.get(k):
            island_occupied+=sum(1 for (y,x) in island_cells if v[k][y][x])

    ent=[]                      # ground entities: chaser + zoner + dps
    for k in ('chaser','zoner','dps'):
        if v.get(k): ent += comps(v[k],1,diag=True)
    eq=quad(ent,h,w)
    sq=quad(comps(v['static'],1,diag=True) if v.get('static') else [],h,w)
    aq=quad(comps(v['mobAir'],1,diag=True) if v.get('mobAir') else [],h,w)

    counts={k:len(comps(v[k],1,diag=True)) for k in ('static','chaser','zoner','dps','mobAir') if v.get(k)}
    warn=[f"{x['layer']} {x['placed']}/{x['requested']}" for x in (r.get('warnings') or [])]

    # --- criteria -------------------------------------------------------
    # HARD: a room that trips any of these is thrown away and regenerated.
    # Every one of them is either an explicit instruction or something all 23
    # shipped rooms already satisfy.
    fails=[]
    near_note=None
    if v.get('zoner'):
        bad=[]
        for cells in comps(v['zoner'],1,diag=True):
            ys=[c[0] for c in cells]; xs=[c[1] for c in cells]
            if not (len(cells)==4 and max(ys)-min(ys)==1 and max(xs)-min(xs)==1):
                bad.append(len(cells))
        if bad: fails.append(f"zoner not 2x2: group sizes {bad}")
        # Two blocks with a single cell between them read as one 2x5 wall.
        groups=comps(v['zoner'],1,diag=True)
        near=[]
        for i in range(len(groups)):
            for k in range(i+1,len(groups)):
                dist=min(max(abs(a[0]-b[0]),abs(a[1]-b[1])) for a in groups[i] for b in groups[k])
                if dist<3: near.append(dist)
        near_note = f"zoner blocks {near.count(2)} pair(s) only 1 cell apart" if near else None
        # the generator already enforces these two; verified here so a rule
        # regression cannot slip through into a saved room
        band=w//3                       # visual bottom third (h is visual height)
        low=[(y,x) for y in range(h-band,h) for x in range(len(g[0])) if v['zoner'][y][x]]
        if low: fails.append(f"zoner in the bottom third at {low[:3]}")
    # ORT-128: a spawn layer with zero cells makes every enemy of that role
    # vanish from the room - EncounterActionExecutor.PlaceDirect skips the whole
    # launch when its layer is empty. Fewer cells than enemies degrades
    # gracefully; zero does not.
    empty=[k for k in ('chaser','zoner','dps','mobAir')
           if not v.get(k) or not any(any(r) for r in v[k])]
    if empty: fails.append(f"spawn layer empty: {empty} (ORT-128)")
    # An open wall must not present a one-cell slit: every walkable run along it
    # is at least 2 cells, and at least one run exists. That is all this repo can
    # honestly check - for a fullroom the whole wall line is walkable and *where*
    # along it the game puts the doorway is not ours to know.
    #
    # An earlier version of this rule assumed the doorway sat at width/2 and
    # demanded that its 2x2 be free of static. It produced five false positives on
    # rooms whose wall was 12 cells of clear floor with a static block happening
    # to sit at x=8, and cost two rooms a needless regeneration. Static near an
    # entrance is cover, not a narrow door.
    gate_bad=[]
    gd=data_space_doors(p, r.get('_doors_space','data'))
    raw=r['payload']['ground']; rh,rw=len(raw),len(raw[0])
    for side in ('top','bottom','left','right'):
        if not gd.get(side): continue
        if side in ('top','bottom'):
            yy=0 if side=='top' else rh-1
            walk=[1 if raw[yy][x] else 0 for x in range(rw)]
        else:
            xx=0 if side=='left' else rw-1
            walk=[1 if raw[y][xx] else 0 for y in range(rh)]
        runs=[];cur=0
        for cell in walk+[0]:
            if cell: cur+=1
            else:
                if cur: runs.append(cur)
                cur=0
        if not runs or max(runs)<2 or 1 in runs:
            gate_bad.append(f"{side}{runs}")
    if gate_bad: fails.append(f"doorway narrower than 2 cells at {gate_bad}")
    if spurs: fails.append(f"ground has {spurs} 1-cell spurs")
    if (p.get('stageType') in AIRY_STAGES) and fill > MAX_AIRY_FILL:
        fails.append(f"floor is a solid slab (fill={fill:.2f} > {MAX_AIRY_FILL})")
    # A carved void has to read as a deliberate cut: at least 2 cells thick in
    # both directions. A 2x1 nick at a corner looks like a mistake. The
    # generator aims higher (3 on the long side) but 2x2 is legal - the shipped
    # rooms contain them.
    thin=[]
    for cells in comps(g,0):
        ys=[c[0] for c in cells]; xs=[c[1] for c in cells]
        dy,dx=max(ys)-min(ys)+1, max(xs)-min(xs)+1
        if min(dy,dx)<2: thin.append(f"{dy}x{dx}")
    if thin: fails.append(f"void region too small/thin: {thin}")
    # Every carved void must be a clean rectangle. Calibrated against review:
    # all four rooms accepted by eye had every void region at bbox fill 1.0,
    # and the one rejected for looking "fragmented" had two at 0.67 - corner
    # cuts and centre pits that merged into an L.
    if vrect and min(vrect) < 1.0:
        fails.append(f"void region not a clean rectangle (bbox fill {min(vrect)})")
    # Same-category spawns must never touch, in any of the 8 directions
    # (ORT-93: the game collapses them and crashes). The chaser and dps layers
    # have a relaxed placement pass that drops this constraint when the strict
    # pass cannot fill the stage's count, and since the counts were doubled that
    # pass fires often enough to matter - 25% of pressure and 35% of peak rooms
    # at 16x10.
    for layer in ('chaser','dps','zoner'):
        if not v.get(layer): continue
        touch=[]
        gg=v[layer]; H=len(gg); W=len(gg[0])
        for y in range(H):
            for x in range(W):
                if not gg[y][x]: continue
                for dy,dx in ((0,1),(1,0),(1,1),(1,-1)):
                    ny,nx=y+dy,x+dx
                    if 0<=ny<H and 0<=nx<W and gg[ny][nx]:
                        touch.append(((y,x),(ny,nx)))
        if layer=='zoner':
            continue          # 2x2 blocks touch internally; shape is checked above
        if touch: fails.append(f"{layer} spawns 8-adjacent ({len(touch)} pairs, e.g. {touch[0]})")
    if islands:
        fails.append(f"unreachable ground island(s) {[len(c) for c in islands]} cells"
                     + (f", {island_occupied} entity/static cells stranded" if island_occupied else ""))

    # SOFT: reported, never a rejection. Placement is best-effort (the user's
    # call): a room that cannot hold its stage's counts places what fits and
    # says so. Quadrant balance is advisory too - a room the user accepted in
    # review ran 5:1 across quadrants.
    if sum(eq.values())>=10 and min(eq.values())==0:
        fails.append(f"entity-free quadrant {[k for k,n in eq.items() if n==0]}")
    # Cover has to be spread too: four or more blocks all in one half of the
    # room leaves the other half bare.
    if sum(sq.values())>=4 and sum(1 for n in sq.values() if n)<3:
        fails.append(f"static crowded into fewer than 3 quadrants {sq}")

    # One half of the room must not carry close to twice the other's spawns.
    # The gate was >=12 spawns; lowered to 10 after two rail rooms at 10 and 11
    # spawns came out 9:1 and 8:3 and had to be redone by eye.
    # Calibrated on review: every room accepted by eye sits at or below 1.63,
    # the two rejected for looking one-sided were at 2.4 (left/right) and 1.9
    # (top/bottom). Quadrant ratios do not separate these cases; halves do.
    if sum(eq.values())>=10:
        halves={'left':eq['TL']+eq['BL'], 'right':eq['TR']+eq['BR'],
                'top':eq['TL']+eq['TR'], 'bottom':eq['BL']+eq['BR']}
        for a,b in (('left','right'),('top','bottom')):
            hi,lo=max(halves[a],halves[b]),min(halves[a],halves[b])
            if lo and hi/lo>1.8:
                fails.append(f"{a}/{b} density {halves[a]}:{halves[b]} ({hi/lo:.1f}x)")

    soft=[]
    if near_note: soft.append(near_note)
    if warn: soft.append("shortfall: " + ', '.join(warn))

    if fill<0.5: soft.append(f"sparse ground (fill={fill:.2f})")


    return dict(w=w,h=h,fill=round(fill,2),voidrect=vrect,spurs=spurs,pits=pits,
                sym=(round(sym_lr,2),round(sym_ud,2)),counts=counts,
                eq=eq,sq=sq,aq=aq,warn=warn,fails=fails,soft=soft,lines=render(v),
                missing=sum(x['requested']-x['placed'] for x in (r.get('warnings') or [])),
                ref=os.path.basename(r['_ref']),doors=r['_request']['doors'])

def main(sub='gen'):
    d=os.path.join(SP,sub); rows=[]
    full=open(os.path.join(SP,'report_%s.txt'%sub),'w')
    for f in sorted(os.listdir(d)):
        if not f.endswith('.json'): continue
        r=json.load(open(os.path.join(d,f)))
        if 'error' in r: print(f[:-5],'ERROR',r['error']); continue
        a=analyse(r); key=f[:-5]; rows.append((key,a))
        full.write(f"=== {key}  doors(data)={a['doors']}  ref={a['ref']}  {a['w']}x{a['h']} visual\n")
        for l in a['lines']: full.write('   '+l+'\n')
        full.write(f"  fill={a['fill']} voidrect={a['voidrect']} spurs={a['spurs']} sym={a['sym']}\n")
        full.write(f"  counts={a['counts']}\n  entity_quad={a['eq']} static_quad={a['sq']} air_quad={a['aq']}\n")
        full.write(f"  VERDICT: {'OK' if not a['fails'] else 'REJECT - ' + '; '.join(a['fails'])}\n")
        if a['soft']: full.write("  note: " + '; '.join(a['soft']) + "\n")
        full.write("\n")
    full.close()
    print(f"{'room':14} {'fill':>5} {'spur':>4} {'symLR/UD':>9}  {'S/C/Z/D/A':>14}  verdict")
    for key,a in rows:
        c=a['counts']
        cnt=f"{c.get('static',0)}/{c.get('chaser',0)}/{c.get('zoner',0)}/{c.get('dps',0)}/{c.get('mobAir',0)}"
        print(f"{key:14} {a['fill']:>5} {a['spurs']:>4} {a['sym'][0]:>4}/{a['sym'][1]:<4} {cnt:>14}  "
              + ('OK' if not a['fails'] else 'REJECT: ' + '; '.join(a['fails']))
              + (('  | note: ' + '; '.join(a['soft'])) if a['soft'] else ''))
# Importable: batch.py pulls analyse() from here, and only a direct run should
# render a report (the import used to drop a stray report file next to it).
if __name__ == '__main__':
    main(sys.argv[1] if len(sys.argv) > 1 else 'gen')
