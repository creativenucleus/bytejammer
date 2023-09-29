


























-- ADDED BY TICJAMMER --
__OLDOVR=OVR
OVR=function()
    if __OLDOVR then __OLDOVR() end
    local t=time()
    local a="--[[$DISPLAY_NAME]]--"
    local w=print(a,240,0,0)
    rect(240-w-8,127,w+4,7,1+(t//300)%15)
    print(a,240-w-8+2,128,15-(150+t//380)%15)
end