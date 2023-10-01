CLIENT_ID=--[[$CLIENT_ID]]--
DISPLAY_NAME="--[[$DISPLAY_NAME]]--"

T=0
function TIC()
	cls(1+(CLIENT_ID%15))

    w=print(DISPLAY_NAME,240,0,0,false,2)
    print(DISPLAY_NAME,120-w/2,10,12,false,2)

    text="Jammer " .. CLIENT_ID
    w=print(text,240,0,0,false,4)
    x=120-w/2+20*math.sin(T/80)
    y=64+32*math.sin(T/90)
    print(text,x,y,12,false,4)
    T=T+1
end