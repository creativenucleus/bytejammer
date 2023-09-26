CLIENT=--[[$CLIENT]]--

T=0
function TIC()
	cls(1+(CLIENT%15))
    
    text="Jammer " .. CLIENT
    w=print(text,240,0,0,false,4)
    x=120-w/2+20*math.sin(T/80)
    y=64+32*math.sin(T/90)
    print(text,x,y,12,false,4)
    T=T+1
end