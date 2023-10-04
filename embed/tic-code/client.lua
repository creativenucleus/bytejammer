CLIENT_KEY="--[[$CLIENT_ID]]--"
DISPLAY_NAME="--[[$DISPLAY_NAME]]--"

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