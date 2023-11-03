-- pos: 0,0
function TIC()t=time()//32
for y=0,136 do for x=0,240 do
pix(x,y,(x+y+t)>>3)
end end end



























-- ADDED BY BYTEJAMMER --
__OLDOVR=OVR
OVR=function()
    if __OLDOVR then __OLDOVR() end
    local t=time()
    if t<2000 or t>6000 then return end
    local a="jtruk"
    local w=print(a,240,0,0)
    local y=144-math.sin(math.pi/2*((t-2000)/2000))*17
    rect(240-w-8,y,w+4,7,1+(t//300)%15)
    print(a,240-w-8+2,y+1,15-(150+t//380)%15)
end