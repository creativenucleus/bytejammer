-- pos: 0,0
CLIENT_KEY="Fake machine name"
DISPLAY_NAME="jtruk"

T=0
function TIC()
    local hash=stringToIndex(CLIENT_KEY)
	if hash==12 then hash=0 end
    cls(hash)

    w=print(DISPLAY_NAME,240,0,0,false,2)
    print(DISPLAY_NAME,120-w/2,10,12,false,2)

    w=print(CLIENT_KEY,240,0,0,false,2)
    x=120-w/2+20*math.sin(T/80)
    y=64+32*math.sin(T/90)
    print(CLIENT_KEY,x,y,12,false,2)
    T=T+1
end

function stringToIndex(text)
    local hash=0
    for i=1,#text do
        hash=hash+string.byte(text,i)
    end
    return 1+hash%15
end


























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